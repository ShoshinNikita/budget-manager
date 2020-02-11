// Package pg provides a PostgreSQL implementation for DB
package pg

import (
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
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

// Prepare prepares the database:
//   - create tables
//   - init tables (add days for current month if needed)
//   - run some subproccess
//
func (db *DB) Prepare() error {
	db.log.Debug("create tables")

	// Create tables If Not Exists
	err := createTables(
		db.db,

		&Month{}, &orm.CreateTableOptions{IfNotExists: true},
		&Income{}, &orm.CreateTableOptions{IfNotExists: true},
		&MonthlyPayment{}, &orm.CreateTableOptions{IfNotExists: true},

		&Day{}, &orm.CreateTableOptions{IfNotExists: true},
		&Spend{}, &orm.CreateTableOptions{IfNotExists: true},
		&SpendType{}, &orm.CreateTableOptions{IfNotExists: true},
	)

	err = errors.Wrap(err, "couldn't create tables")
	if err != nil {
		db.log.WithError(err).Error("couldn't create tables")
		return err
	}

	// Init tables if needed
	if err = db.initCurrentMonth(); err != nil {
		return err
	}

	// Init a new month monthly at 00:00
	_, err = db.cron.AddFunc("@monthly", func() {
		err = db.initCurrentMonth()
		if err != nil {
			db.log.WithError(err).Error("couldn't init a new month")
		}
	})
	if err != nil {
		db.log.WithError(err).Error("can't add function to cron")
		return errors.Wrap(err, "could't add function to cron")
	}

	// Start cron
	db.cron.Start()

	return nil
}

// DropDB drops all tables and relations. USE ONLY IN TESTS!
func (db *DB) DropDB() error {
	return dropTables(db.db,
		&Month{}, &orm.DropTableOptions{IfExists: true},
		&Income{}, &orm.DropTableOptions{IfExists: true},
		&MonthlyPayment{}, &orm.DropTableOptions{IfExists: true},

		&Day{}, &orm.DropTableOptions{IfExists: true},
		&Spend{}, &orm.DropTableOptions{IfExists: true},
		&SpendType{}, &orm.DropTableOptions{IfExists: true},
	)
}

func (db *DB) initCurrentMonth() error {
	now := time.Now()
	year, month, _ := now.Date()

	err := db.db.Model(&Month{}).
		Column("id").
		Where("year = ? AND month = ?", year, month).
		Select()
	if err == nil {
		return nil
	}
	if err != pg.ErrNoRows {
		db.log.WithError(err).Error("couldn't select the current month")
		return err
	}

	// We have to init the current month
	log := db.log

	// Add current month
	log.Debug("init the current month")

	currentMonth := &Month{Year: year, Month: month}
	if err = db.db.Insert(currentMonth); err != nil {
		log.WithError(err).Error("could't init the current month")
		return errors.Wrap(err, "could't init the current month")
	}

	monthID := currentMonth.ID
	log = log.WithField("month_id", monthID)
	log.Debug("current month was successfully inited")

	// Add days for the current month
	daysNumber := daysInMonth(now)
	days := make([]*Day, daysNumber)
	for i := range days {
		days[i] = &Day{MonthID: monthID, Day: i + 1, Saldo: 0}
	}

	if err = db.db.Insert(&days); err != nil {
		log.WithError(err).Error("couldn't insert days for current month")
		return errors.Wrap(err, "couldn't insert days for current month")
	}

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
