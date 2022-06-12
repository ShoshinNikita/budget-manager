package app

import (
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

type App struct {
	config  Config
	version string
	gitHash string

	db     Database
	log    logger.Logger
	server *web.Server

	shutdownSignal chan struct{}
}

type Database interface {
	Shutdown() error

	web.Database
}

// NewApp returns a new instance of App
func NewApp(cfg Config, log logger.Logger, version, gitHash string) *App {
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
	if err := app.prepareWebServer(); err != nil {
		return errors.Wrap(err, "couldn't prepare web server")
	}

	return nil
}

func (app *App) prepareDB() (err error) {
	return nil
}

//nolint:unparam
func (app *App) prepareWebServer() error {
	app.server = web.NewServer(app.config.Server, app.db, app.log, app.version, app.gitHash)
	return nil
}

// Run runs web server. This method should be called in a goroutine
func (app *App) Run() error {
	app.log.WithFields(logger.Fields{
		"version":  app.version,
		"git_hash": app.gitHash,
	}).Info("start app")

	errCh := make(chan error, 1)
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
