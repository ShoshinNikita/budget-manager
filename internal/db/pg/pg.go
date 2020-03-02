// Package pg provides a PostgreSQL implementation for DB
package pg

import (
	"strconv"
	"time"

	"github.com/go-pg/migrations/v7"
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	pg_migrations "github.com/ShoshinNikita/budget-manager/internal/db/pg/migrations"
)

const (
	connectRetries      = 5
	connectRetryTimeout = 2 * time.Second
)

type Config struct {
	Host     string `env:"DB_PG_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PG_PORT" envDefault:"5432"`
	User     string `env:"DB_PG_USER" envDefault:"postgres"`
	Password string `env:"DB_PG_PASSWORD"`
	Database string `env:"DB_PG_DATABASE" envDefault:"postgres"`
}

type DB struct {
	db   *pg.DB
	cron *cron.Cron
	log  logrus.FieldLogger
}

// NewDB creates a new connection to the db and pings it
func NewDB(config Config, log logrus.FieldLogger) (*DB, error) {
	db := &DB{
		log: log.WithField("db_type", "pg"),
		db: pg.Connect(&pg.Options{
			Addr:     config.Host + ":" + strconv.Itoa(config.Port),
			User:     config.User,
			Password: config.Password,
			Database: config.Database,
		}),
		cron: cron.New(
			cron.WithLogger(cronLogger{log: log.WithField("component", "cron")}),
		),
	}

	// Try to ping the DB
	for i := 0; i < connectRetries; i++ {
		log := db.log.WithField("try", i+1)
		if i != 0 {
			log.Debug("ping DB")
		}

		if ping(db.db) {
			break
		}
		log.Debug("couldn't ping DB")
		if i+1 == connectRetries {
			// Don't sleep extra time
			return nil, errors.New("database is down")
		}

		time.Sleep(connectRetryTimeout)
	}

	return db, nil
}

const migrationTable = "migrations"

// Prepare prepares the database:
//   - create tables
//   - init tables (add days for current month if needed)
//   - run some subproccess
//
func (db *DB) Prepare() error {
	// Create a new migrator
	migrator := migrations.NewCollection().SetTableName(migrationTable).DisableSQLAutodiscover(true)

	// Register migrations
	pg_migrations.RegisterMigrations(migrator)
	if len(migrator.Migrations()) != pg_migrations.MigrationNumber {
		const msg = "invalid number of registered migrations"
		db.log.WithFields(logrus.Fields{
			"got": len(migrator.Migrations()), "want": pg_migrations.MigrationNumber,
		}).Error(msg)
		return errors.New(msg)
	}

	// Init migration table
	if _, _, err := migrator.Run(db.db, "init"); err != nil {
		const msg = "couldn't init migration table"
		db.log.WithError(err).Error(err)
		return errors.Wrap(err, msg)
	}

	// Run migrations
	db.log.Debug("run migrations")
	oldVersion, newVersion, err := migrator.Run(db.db, "up")
	if err != nil {
		const msg = "couldn't run migrations"
		db.log.WithError(err).Error(err)
		return errors.Wrap(err, msg)
	}

	db.log.WithFields(logrus.Fields{
		"old_version": oldVersion, "new_version": newVersion,
	}).Info("migration process is finished")

	// Init tables if needed
	if err = db.initCurrentMonth(); err != nil {
		const msg = "couldn't init the current month"
		db.log.WithError(err).Error(msg)
		return errors.Wrap(err, msg)
	}

	// Init a new month monthly at 00:00
	_, err = db.cron.AddFunc("@monthly", func() {
		err = db.initCurrentMonth()
		if err != nil {
			db.log.WithError(err).Error("couldn't init a new month")
		}
	})
	if err != nil {
		const msg = "couldn't add a function to cron"
		db.log.WithError(err).Error(msg)
		return errors.Wrap(err, msg)
	}

	// Start cron
	db.cron.Start()

	return nil
}

// DropDB drops all tables and relations. USE ONLY IN TESTS!
func (db *DB) DropDB() error {
	_, err := db.db.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
		GRANT ALL ON SCHEMA public TO public;
	`)
	return err
}

func (db *DB) initCurrentMonth() error {
	year, month, _ := time.Now().Date()
	return db.addMonth(year, month)
}

func (db *DB) addMonth(year int, month time.Month) error {
	err := db.db.Model(&Month{}).
		Column("id").
		Where("year = ? AND month = ?", year, month).
		Select()
	if err == nil {
		// The month is already created
		return nil
	}
	if err != pg.ErrNoRows {
		// Unexpected error
		db.log.WithError(err).Error("unexpected error: couldn't select the current month")
		return err
	}

	// We have to init the current month
	log := db.log

	// Add the current month
	log.Debug("init the current month")

	currentMonth := &Month{Year: year, Month: month}
	if err = db.db.Insert(currentMonth); err != nil {
		const msg = "couldn't init the current month"
		log.WithError(err).Error(msg)
		return errors.Wrap(err, msg)
	}

	monthID := currentMonth.ID
	log = log.WithField("month_id", monthID)
	log.Debug("current month was successfully inited")

	// Add days for the current month
	log.Debug("init days of the current month")

	daysNumber := daysInMonth(year, month)
	days := make([]*Day, daysNumber)
	for i := range days {
		days[i] = &Day{MonthID: monthID, Day: i + 1, Saldo: 0}
	}

	if err = db.db.Insert(&days); err != nil {
		log.WithError(err).Error("couldn't insert days for current month")
		return errors.Wrap(err, "couldn't insert days for current month")
	}
	log.Debug("days of the current month was successfully inited")

	return nil
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	// cron
	db.log.Debug("wait to stop cron jobs")
	ctx := db.cron.Stop()
	<-ctx.Done()
	db.log.Debug("all cron jobs were stopped")

	// Database connection
	db.log.Debug("close database connection")
	return db.db.Close()
}
