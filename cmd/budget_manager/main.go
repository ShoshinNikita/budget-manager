package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/logger"
	"github.com/ShoshinNikita/budget_manager/internal/web"
)

type Config struct {
	// Is debug mode on
	Debug bool `env:"DEBUG" envDefault:"false"`

	Logger logger.Config
	DB     db.Config
	Server web.Config
}

func main() {
	// Parse config
	cnf, err := parseConfig()
	if err != nil {
		log.Fatalf("can't parse config: %s", err)
	}

	log := logger.New(cnf.Logger, cnf.Debug)
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

	server := web.NewServer(cnf.Server, db, log)
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

func parseConfig() (*Config, error) {
	cnf := &Config{}
	if err := env.Parse(cnf); err != nil {
		return nil, err
	}

	return cnf, nil
}
