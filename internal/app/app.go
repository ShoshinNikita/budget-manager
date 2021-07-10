package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type App struct {
	config  Config
	version string
	gitHash string

	db     Database
	log    logrus.FieldLogger
	server *web.Server
}

type Database interface {
	Prepare() error
	Shutdown() error

	web.Database
}

// NewApp returns a new instance of App
func NewApp(cfg Config, version, gitHash string) *App {
	return &App{
		config:  cfg,
		version: version,
		gitHash: gitHash,
	}
}

// PrepareComponents prepares logger, db and web server
func (app *App) PrepareComponents() error {
	// Logger
	app.prepareLogger()
	app.log.Info("logger is initialized")

	// DB
	app.log.Info("prepare database")
	if err := app.prepareDB(); err != nil {
		return errors.Wrap(err, "database init error")
	}
	app.log.Info("database is initialized")

	// Web Server
	app.log.Info("prepare web server")
	app.prepareWebServer()
	app.log.Info("web server is initialized")

	return nil
}

func (app *App) prepareLogger() {
	app.log = logger.New(app.config.Logger)
}

func (app *App) prepareDB() (err error) {
	// Connect
	app.log.Debug("connect to the db")

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

	app.log.Debug("connection is ready")

	// Prepare the db
	app.log.Debug("prepare db")
	err = app.db.Prepare()
	if err != nil {
		return errors.Wrap(err, "couldn't prepare the db")
	}
	app.log.Debug("preparations were successful")

	return nil
}

func (app *App) prepareWebServer() {
	app.server = web.NewServer(
		app.config.Server, app.db, app.log, app.version, app.gitHash,
	)
	app.server.Prepare()
}

// Run runs web server and waits for a server error or an interrupt signal
func (app *App) Run() (appErr error) {
	app.log.WithFields(logrus.Fields{
		"version":  app.version,
		"git_hash": app.gitHash,
	}).Info("start app")

	// Start the application
	webServerError := make(chan error, 1)
	go func() {
		webServerError <- app.server.ListenAndServer()
	}()

	// Wait for interrupt signal
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-term:
		app.log.Warn("got an interrupt signal")
	case err := <-webServerError:
		appErr = err
		app.log.WithError(err).Warn("server is down")
	}

	app.Shutdown()

	return appErr
}

// Shutdown shutdowns the app components
func (app *App) Shutdown() {
	app.log.Info("shutdown components")

	// Server
	app.log.Info("shutdown web server")
	err := app.server.Shutdown()
	if err != nil {
		app.log.WithError(err).Error("couldn't shutdown the server gracefully")
	}

	// Database
	app.log.Info("shutdown the database")
	err = app.db.Shutdown()
	if err != nil {
		app.log.WithError(err).Error("couldn't shutdown the db gracefully")
	}

	app.log.Info("shutdowns are completed")
}
