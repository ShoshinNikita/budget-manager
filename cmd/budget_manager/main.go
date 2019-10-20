package main

import (
	"os"
	"os/signal"
	"syscall"

	clog "github.com/ShoshinNikita/go-clog/v3"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/web"
)

func main() {
	log := clog.NewDevLogger()
	log.Info("start")

	// Connect to the db
	log.Info("connect to the db")

	dbOpts := db.NewDBOptions{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Database: "postgres",
	}
	db, err := db.NewDB(dbOpts, log)
	if err != nil {
		log.Fatal("couldn't connect to the db", "error", err)
	}

	log.Info("connection was successful")

	// Prepare the db
	log.Info("prepare db")

	err = db.Prepare()
	if err != nil {
		log.Fatal("couldn't prepare the db", "error", err)
	}

	log.Info("preparations were successful")

	// Create a new server instance
	log.Info("create Server instance")

	serverOpts := web.NewServerOptions{
		Port: ":8080",
	}
	server := web.NewServer(serverOpts, db, log)
	server.Prepare()

	// Start server
	serverError := make(chan struct{})
	go func() {
		log.Info("start Server")
		if err := server.ListenAndServer(); err != nil {
			log.Errorf("server fatal error: %s\n", err)
			close(serverError)
		}
	}()

	// Wait for interrupt signal
	term := make(chan os.Signal)
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
		log.Errorf("can't shutdown the db gracefully: %s\n", err)
	}

	log.Info("shutdowns are completed")
	log.Info("stop")
}
