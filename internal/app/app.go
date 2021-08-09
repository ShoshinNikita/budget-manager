package app

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type App struct {
	config  Config
	version string
	gitHash string

	db     Database
	log    logrus.FieldLogger
	server *web.Server

	shutdownSignal chan struct{}
}

type Database interface {
	Prepare() error
	InitMonth(ctx context.Context, year int, month time.Month) error
	Shutdown() error

	web.Database
}

// NewApp returns a new instance of App
func NewApp(cfg Config, log *logrus.Logger, version, gitHash string) *App {
	return &App{
		config:  cfg,
		version: version,
		gitHash: gitHash,
		//
		log: log,
		//
		shutdownSignal: make(chan struct{}),
	}
}

// PrepareComponents prepares logger, db and web server
func (app *App) PrepareComponents() error {
	app.log.Debug("prepare database")
	if err := app.prepareDB(); err != nil {
		return errors.Wrap(err, "couldn't prepare database")
	}

	app.log.Debug("prepare web server")
	app.prepareWebServer()

	return nil
}

func (app *App) prepareDB() (err error) {
	// Connect
	switch app.config.DBType {
	case "postgres", "postgresql":
		app.log.Debug("db type is PostgreSQL")
		app.db, err = pg.NewDB(app.config.PostgresDB, app.log)
	default:
		err = errors.New("unsupported DB type")
	}
	if err != nil {
		return errors.Wrap(err, "couldn't create DB connection")
	}

	// Prepare the db
	if err := app.db.Prepare(); err != nil {
		return errors.Wrap(err, "couldn't prepare the db")
	}

	// Init the current month
	if err := app.initMonth(time.Now()); err != nil {
		return errors.Wrap(err, "couldn't init the current month")
	}

	return nil
}

func (app *App) prepareWebServer() {
	app.server = web.NewServer(
		app.config.Server, app.db, app.log, app.version, app.gitHash,
	)
	app.server.Prepare()
}

// Run runs web server. This method should be called in a goroutine
func (app *App) Run() error {
	app.log.WithFields(logrus.Fields{
		"version":  app.version,
		"git_hash": app.gitHash,
	}).Info("start app")

	errCh := make(chan error, 2)
	startBackroundJob := func(errorMsg string, f func() error) {
		go func() {
			err := f()
			if err != nil {
				app.log.WithError(err).Error(errorMsg)
			}
			errCh <- err
		}()
	}
	startBackroundJob("web server failed", app.server.ListenAndServer)
	startBackroundJob("month init failed", app.startMonthInit)

	return <-errCh
}

// Shutdown shutdowns the app components
func (app *App) Shutdown() {
	app.log.Info("shutdown app")
	close(app.shutdownSignal)

	app.log.Debug("shutdown web server")
	if err := app.server.Shutdown(); err != nil {
		app.log.WithError(err).Error("couldn't shutdown the server gracefully")
	}

	app.log.Debug("shutdown the database")
	if err := app.db.Shutdown(); err != nil {
		app.log.WithError(err).Error("couldn't shutdown the db gracefully")
	}
}

func (app *App) startMonthInit() error {
	for {
		after := calculateTimeToNextMonthInit(time.Now())

		select {
		case now := <-time.After(after):
			app.log.WithField("date", now.Format("2006-01-02")).Debug("init a new month")

			if err := app.initMonth(now); err != nil {
				return errors.Wrap(err, "couldn't init a new month")
			}

		case <-app.shutdownSignal:
			return nil
		}
	}
}

// calculateTimeToNextMonthInit returns time left to the start (00:00) of the next month
func calculateTimeToNextMonthInit(now time.Time) time.Duration {
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return nextMonth.Sub(now)
}

// initMonth inits month for the passed date
func (app *App) initMonth(t time.Time) error {
	year, month, _ := t.Date()
	return app.db.InitMonth(context.Background(), year, month)
}
