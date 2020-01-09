package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShoshinNikita/go-clog/v3"
	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

func main() {
	// Create a new application
	app := NewApp()

	// Parse application config
	if err := app.ParseConfig(); err != nil {
		log.Fatalln(err)
	}

	// Prepare the application
	if err := app.PrepareComponents(); err != nil {
		log.Fatalln(err)
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}

type App struct {
	config Config

	db     *db.DB
	log    *clog.Logger
	server *web.Server
}

type Config struct {
	Debug bool `env:"DEBUG" envDefault:"false"`

	Logger logger.Config
	DB     db.Config
	Server web.Config
}

// NewApp returns a new instance of App
func NewApp() *App {
	return &App{}
}

// ParseConfig parses app config
func (app *App) ParseConfig() error {
	if err := env.Parse(&app.config); err != nil {
		return err
	}
	return nil
}

// PrepareComponents prepares logger, db and web server
func (app *App) PrepareComponents() error {
	// Logger
	app.prepareLogger()
	app.log.Info("logger is initialized")

	// DB
	app.log.Info("prepare database...")
	if err := app.prepareDB(); err != nil {
		return errors.Wrap(err, "database init error")
	}
	app.log.Info("database is initialized")

	// Web Server
	app.log.Info("prepare web server...")
	app.prepareWebServer()
	app.log.Info("web server is initialized")

	return nil
}

func (app *App) prepareLogger() {
	app.log = logger.New(app.config.Logger, app.config.Debug)
}

func (app *App) prepareDB() (err error) {
	// Connect
	app.log.Debug("connect to the db...")
	app.db, err = db.NewDB(app.config.DB, app.log.WithPrefix("[database]"))
	if err != nil {
		return errors.Wrap(err, "couldn't connect to the db")
	}
	app.log.Debug("connection is ready")

	// Prepare the db
	app.log.Debug("prepare db...")
	err = app.db.Prepare()
	if err != nil {
		return errors.Wrap(err, "couldn't prepare the db")
	}
	app.log.Debug("preparations were successful")

	return nil
}

func (app *App) prepareWebServer() {
	app.server = web.NewServer(
		app.config.Server, app.db, app.log.WithPrefix("[server]"), app.config.Debug,
	)
	app.server.Prepare()
}

// Run runs web server and waits for a server error or an interrupt signal
func (app *App) Run() (appErr error) {
	app.log.Info("start app")

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
		app.log.Warnf("server is down: %s", err)
	}

	app.Shutdown()

	return appErr
}

// Shutdown shutdowns the app components
func (app *App) Shutdown() {
	app.log.Info("shutdown components...")

	// Server
	app.log.Info("shutdown web server")
	err := app.server.Shutdown()
	if err != nil {
		app.log.Errorf("can't shutdown the db gracefully: %s", err)
	}

	// Database
	app.log.Info("shutdown the database")
	err = app.db.Shutdown()
	if err != nil {
		app.log.Errorf("can't shutdown the db gracefully: %s", err)
	}

	app.log.Info("shutdowns are completed")
}