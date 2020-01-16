package pg

import (
	"strconv"
	"time"

	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
)

const (
	connectRetries      = 5
	connectRetryTimeout = 2 * time.Second
)

type Config struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD"`
	Database string `env:"DB_DATABASE" envDefault:"postgres"`
}

type DB struct {
	db   *pg.DB
	cron *cron.Cron
	log  *clog.Logger
}

// NewDB creates a new connection to the db and pings it
func NewDB(config Config, log *clog.Logger) (*DB, error) {
	db := pg.Connect(&pg.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
	})

	// Try to ping the DB
	for i := 0; i < connectRetries; i++ {
		if i != 0 {
			log.Debugf("ping DB, try #%d", i+1)
		}

		if ping(db) {
			break
		}
		log.Debug("couldn't ping DB")
		if i+1 == connectRetries {
			// Don't sleep extra time
			return nil, errors.New("database is down")
		}

		time.Sleep(connectRetryTimeout)
	}

	cron := cron.New(cron.WithLogger(cronLogger{log: log.WithPrefix("[cron]")}))

	return &DB{db: db, log: log, cron: cron}, nil
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
		db.log.Error(err)
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
			db.log.Errorf("can't init the new month: %v", err)
		}
	})
	if err != nil {
		err = errors.Wrap(err, "can't add function to cron")
		db.log.Error(err)
		return err
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

	if err == pg.ErrNoRows {
		// We have to init the current month

		// Add current month
		db.log.Debug("init the current month")

		currentMonth := &Month{Year: year, Month: month}
		err = db.db.Insert(currentMonth)
		if err != nil {
			err = errors.Wrap(err, "can't init current month")
			db.log.Error(err)
			return err
		}

		monthID := currentMonth.ID
		db.log.Debugf("current month id: '%d'", monthID)

		// Add days for the current month
		daysNumber := daysInMonth(now)
		days := make([]*Day, daysNumber)
		for i := range days {
			days[i] = &Day{MonthID: monthID, Day: i + 1, Saldo: 0}
		}

		err = db.db.Insert(&days)
		if err != nil {
			err = errors.Wrap(err, "can't insert days for current month")
			db.log.Error(err)
			return err
		}
	}

	return nil
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	// cron
	db.log.Debug("wait to stop cron jobs...")
	ctx := db.cron.Stop()
	<-ctx.Done()
	db.log.Debug("all cron jobs were stopped")

	// Database connection
	db.log.Debug("close database connection")
	return db.db.Close()
}
