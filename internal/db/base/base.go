// package base provides a PostgreSQL implementation for DB
package base

import (
	"context"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/types"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

const (
	connectRetries      = 10
	connectRetryTimeout = 500 * time.Millisecond
)

const (
	Question = sqlx.Question
	Dollar   = sqlx.Dollar
)

type DB struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewDB creates a new connection to the db and pings it
func NewDB(driverName string, dataSourceName string, placeholder sqlx.Placeholder, log logger.Logger) (*DB, error) {
	conn, err := sqlx.Open(driverName, dataSourceName, placeholder, log)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open a connection to db")
	}
	db := &DB{
		log: log.WithField("db_type", "pg"),
		db:  conn,
	}

	// Try to ping the DB
	for i := 0; i < connectRetries; i++ {
		err := db.db.Ping(context.Background())
		if err == nil {
			break
		}

		log.WithError(err).WithField("try", i+1).Debug("couldn't ping DB")
		if i+1 == connectRetries {
			// Don't sleep extra time
			return nil, errors.New("database is down")
		}

		time.Sleep(connectRetryTimeout)
	}

	return db, nil
}

// Prepare runs the migrations
func (db *DB) Prepare() error {
	migrator, err := migrations.NewMigrator(db.log)
	if err != nil {
		return errors.Wrap(err, "couldn't prepare migrator")
	}

	if err := migrator.Migrate(db.db.GetInternalDB()); err != nil {
		return errors.Wrap(err, "couldn't apply migrations")
	}

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
	err := db.db.RunInTransaction(context.Background(), func(tx *sqlx.Tx) error {
		return tx.Get(&n, `SELECT COUNT(DISTINCT table_name) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = 'public'`)
	})
	if err != nil {
		return errors.Wrap(err, "couldn't get number of tables")
	}
	if n != tableNumber {
		return errors.Errorf("invalid number of tables: '%d', expected: '%d'", n, tableNumber)
	}

	type column struct {
		Name    string       `db:"column_name"`
		Type    string       `db:"data_type"`
		IsNull  bool         `db:"is_nullable"`
		Default types.String `db:"column_default"`
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
		// Skip table for migrations
	}
	for _, table := range tables {
		var columnsInDB []column
		err := db.db.RunInTransaction(context.Background(), func(tx *sqlx.Tx) error {
			return tx.Select(
				&columnsInDB,
				`SELECT column_name, data_type, is_nullable::bool, column_default
				FROM INFORMATION_SCHEMA.COLUMNS
				WHERE table_name = ?
				ORDER BY column_name`,
				table.name,
			)
		})
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
	return db.db.Close()
}
