package main

import (
	"os"
	"os/signal"
	"syscall"

	clog "github.com/ShoshinNikita/go-clog/v3"

	"github.com/ShoshinNikita/budget_manager/internal/db"
)

func main() {
	log := clog.NewDevLogger()
	log.Info("start")

	// Connect to the db
	log.Info("connect to the db")

	opts := db.NewDBOptions{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Database: "postgres",
	}
	db, err := db.NewDB(opts, log)
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

	shutdowned := make(chan struct{})
	go func() {
		term := make(chan os.Signal)
		signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
		<-term

		log.Warn("got an interrupt signal")
		log.Warn("shutdown services")

		log.Info("shutdown the database")
		err = db.Shutdown()
		if err != nil {
			log.Errorf("can't shutdown the db gracefully: %s\n", err)
		}

		log.Info("shutdowns are completed")
		close(shutdowned)
	}()

	<-shutdowned
	log.Info("stop")
}
