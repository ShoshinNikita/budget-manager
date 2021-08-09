// Package pg provides a PostgreSQL implementation for DB
package pg

import (
	"context"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg/migrations"
)

const (
	connectRetries      = 10
	connectRetryTimeout = 250 * time.Millisecond
)

type Config struct {
	Host     string `env:"DB_PG_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PG_PORT" envDefault:"5432"`
	User     string `env:"DB_PG_USER" envDefault:"postgres"`
	Password string `env:"DB_PG_PASSWORD"`
	Database string `env:"DB_PG_DATABASE" envDefault:"postgres"`
}

type DB struct {
	db  *pg.DB
	log logrus.FieldLogger
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
	}

	// Try to ping the DB
	for i := 0; i < connectRetries; i++ {
		log := db.log.WithField("try", i+1)
		if i != 0 {
			log.Debug("ping DB")
		}

		err := db.db.Ping(context.Background())
		if err == nil {
			break
		}
		log.WithError(err).Debug("couldn't ping DB")
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
	// Create a new migrator
	migrator := migrations.NewMigrator()

	// Check number of migrations
	if len(migrator.Migrations()) != migrations.MigrationNumber {
		return errors.Errorf("invalid number of registered migrations: %d (want %d)",
			len(migrator.Migrations()), migrations.MigrationNumber)
	}

	// Init migration table
	if _, _, err := migrator.Run(db.db, "init"); err != nil {
		return errors.Wrap(err, "couldn't init migration table")
	}

	// Run migrations
	db.log.Debug("run migrations")
	oldVersion, newVersion, err := migrator.Run(db.db, "up")
	if err != nil {
		return errors.Wrap(err, "couldn't run migrations")
	}

	db.log.WithFields(logrus.Fields{
		"old_version": oldVersion, "new_version": newVersion,
	}).Debug("migration process was finished")

	// Check the tables
	if err := db.checkCreatedTables(); err != nil {
		return errors.Wrap(err, "database schema is invalid")
	}

	return nil
}

// checkCreatedTables checks tables and their descriptions
//
//nolint:funlen
func (db *DB) checkCreatedTables() error {
	const tableNumber = 7
	var n int
	_, err := db.db.Query(pg.Scan(&n),
		`SELECT COUNT(DISTINCT table_name) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = 'public'`,
	)
	if err != nil {
		return errors.Wrap(err, "couldn't get number of tables")
	}
	if n != tableNumber {
		return errors.Errorf("invalid number of tables: '%d', expected: '%d'", n, tableNumber)
	}

	type column struct {
		Name    string `pg:"column_name"`
		Type    string `pg:"data_type"`
		IsNull  bool   `pg:"is_nullable"`
		Default string `pg:"column_default"`
	}
	tables := []struct {
		name    string
		columns []column
	}{
		// Columns must be sorted by name
		{
			name: "months",
			columns: []column{
				{Name: "daily_budget", Type: "bigint", Default: "0"},
				{Name: "id", Type: "bigint", Default: "nextval('months_id_seq'::regclass)"},
				{Name: "month", Type: "bigint"},
				{Name: "result", Type: "bigint", Default: "0"},
				{Name: "total_income", Type: "bigint", Default: "0"},
				{Name: "total_spend", Type: "bigint", Default: "0"},
				{Name: "year", Type: "bigint"},
			},
		},
		{
			name: "days",
			columns: []column{
				{Name: "day", Type: "bigint"},
				{Name: "id", Type: "bigint", Default: "nextval('days_id_seq'::regclass)"},
				{Name: "month_id", Type: "bigint"},
				{Name: "saldo", Type: "bigint", Default: "0"},
			},
		},
		{
			name: "incomes",
			columns: []column{
				{Name: "id", Type: "bigint", Default: "nextval('incomes_id_seq'::regclass)"},
				{Name: "income", Type: "bigint"},
				{Name: "month_id", Type: "bigint"},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text"},
			},
		},
		{
			name: "monthly_payments",
			columns: []column{
				{Name: "cost", Type: "bigint"},
				{Name: "id", Type: "bigint", Default: "nextval('monthly_payments_id_seq'::regclass)"},
				{Name: "month_id", Type: "bigint"},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text"},
				{Name: "type_id", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "spends",
			columns: []column{
				{Name: "cost", Type: "bigint"},
				{Name: "day_id", Type: "bigint"},
				{Name: "id", Type: "bigint", Default: "nextval('spends_id_seq'::regclass)"},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text"},
				{Name: "type_id", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "spend_types",
			columns: []column{
				{Name: "id", Type: "bigint", Default: "nextval('spend_types_id_seq'::regclass)"},
				{Name: "name", Type: "text"},
				{Name: "parent_id", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "migrations",
			columns: []column{
				{Name: "created_at", Type: "timestamp with time zone", IsNull: true},
				{Name: "id", Type: "integer", Default: "nextval('migrations_id_seq'::regclass)"},
				{Name: "version", Type: "bigint", IsNull: true},
			},
		},
	}
	var columnsInDB []column
	for _, table := range tables {
		_, err := db.db.Query(&columnsInDB,
			`SELECT column_name, data_type, is_nullable::bool , column_default
			   FROM INFORMATION_SCHEMA.COLUMNS
			  WHERE table_name = ?
			  ORDER BY column_name`, table.name,
		)
		if err != nil {
			return errors.Wrapf(err, "couldn't get description of table '%s'", table.name)
		}

		err = errors.Errorf("table '%s' has wrong columns: '%+v', expected: '%+v'", table.name, columnsInDB, table.columns)
		if len(table.columns) != len(columnsInDB) {
			return err
		}
		for i := range table.columns {
			if table.columns[i] != columnsInDB[i] {
				return err
			}
		}
	}
	return nil
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	db.log.Debug("close database connection")
	return db.db.Close()
}
