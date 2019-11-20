package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/caarlos0/env/v6"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/web"
)

type Config struct {
	// Is debug mode on
	Debug bool `env:"DEBUG" envDefault:"false"`

	Logger struct {
		// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
		// It is always debug, when debug mode is on
		Level string `env:"LOGGER_LEVEL" envDefault:"info"`
	}

	DB db.Config

	Server struct {
		Port string `env:"SERVER_PORT" envDefault:":8080"`
	}
}

func main() {
	// Parse config
	var cnf Config
	if err := env.Parse(&cnf); err != nil {
		log.Fatalf("can't parse config: %s", err)
	}

	// Setup logger. Use prod config by default
	log := clog.NewProdConfig().SetLevel(logLevelFromString(cnf.Logger.Level)).Build()
	if cnf.Debug {
		log = clog.NewDevConfig().SetLevel(clog.LevelDebug).Build()
	}

	log.Info("start")

	// Connect to the db
	log.Info("connect to the db")

	db, err := db.NewDB(cnf.DB, log)
	if err != nil {
		log.Fatalf("couldn't connect to the db: %s", err)
	}

	log.Info("connection was successful")

	// Prepare the db
	log.Info("prepare db")

	err = db.Prepare()
	if err != nil {
		log.Fatalf("couldn't prepare the db: %s", err)
	}

	log.Info("preparations were successful")

	// Create a new server instance
	log.Info("create Server instance")

	serverOpts := web.NewServerOptions{
		Port: cnf.Server.Port,
	}
	server := web.NewServer(serverOpts, db, log)
	server.Prepare()

	// Start server
	serverError := make(chan struct{})
	go func() {
		log.Info("start Server")
		if err := server.ListenAndServer(); err != nil {
			log.Errorf("server fatal error: %s", err)
			close(serverError)
		}
	}()

	// Wait for interrupt signal
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-term:
		log.Warn("got an interrupt signal")
	case <-serverError:
		log.Warn("server is down")
	}

	log.Warn("shutdown services")

	// Server
	log.Info("shutdown the server")
	err = server.Shutdown()
	if err != nil {
		log.Errorf("can't shutdown the db gracefully: %s", err)
	}

	// Database
	log.Info("shutdown the database")
	err = db.Shutdown()
	if err != nil {
		log.Errorf("can't shutdown the db gracefully: %s", err)
	}

	log.Info("shutdowns are completed")
	log.Info("stop")
}

func logLevelFromString(lvl string) clog.LogLevel {
	switch lvl {
	case "dbg", "debug":
		return clog.LevelDebug
	case "inf", "info":
		return clog.LevelInfo
	case "warn", "warning":
		return clog.LevelWarn
	case "err", "error":
		return clog.LevelError
	case "fatal":
		return clog.LevelFatal
	default:
		return clog.LevelInfo
	}
}
