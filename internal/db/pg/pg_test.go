// +build integration

package pg

import (
	"io/ioutil"
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCreatedTables(t *testing.T) {
	t.Run("run migrations against the empty db", func(t *testing.T) {
		db := initDB(t)
		checkTables(t, db)
		cleanUp(t, db)
	})

	t.Run("run migrations against the up-to-date db", func(t *testing.T) {
		db := initDB(t)
		checkTables(t, db)
		require.Nil(t, db.Shutdown())

		db = initDB(t)
		checkTables(t, db)
		cleanUp(t, db)
	})
}

// nolint:funlen
func checkTables(t *testing.T, db *DB) {
	// Check table number at first
	ok := t.Run("check table number", func(t *testing.T) {
		const tableNumber = 7

		var n int
		_, err := db.db.Query(pg.Scan(&n),
			`SELECT COUNT(DISTINCT table_name) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = 'public'`,
		)
		require.Nil(t, err)
		require.Equal(t, tableNumber, n)
	})
	if !ok {
		t.FailNow()
	}

	type column struct {
		Name    string `pg:"column_name"`
		Type    string `pg:"data_type"`
		IsNull  bool   `pg:"is_nullable"`
		Default string `pg:"column_default"`
	}
	tables := []struct {
		name        string
		description []column
	}{
		// Columns must be sorted by name
		{
			name: "months",
			description: []column{
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
			description: []column{
				{Name: "day", Type: "bigint"},
				{Name: "id", Type: "bigint", Default: "nextval('days_id_seq'::regclass)"},
				{Name: "month_id", Type: "bigint"},
				{Name: "saldo", Type: "bigint", Default: "0"},
			},
		},
		{
			name: "incomes",
			description: []column{
				{Name: "id", Type: "bigint", Default: "nextval('incomes_id_seq'::regclass)"},
				{Name: "income", Type: "bigint"},
				{Name: "month_id", Type: "bigint"},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text"},
			},
		},
		{
			name: "monthly_payments",
			description: []column{
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
			description: []column{
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
			description: []column{
				{Name: "id", Type: "bigint", Default: "nextval('spend_types_id_seq'::regclass)"},
				{Name: "name", Type: "text"},
			},
		},
		{
			name: "migrations",
			description: []column{
				{Name: "created_at", Type: "timestamp with time zone", IsNull: true},
				{Name: "id", Type: "integer", Default: "nextval('migrations_id_seq'::regclass)"},
				{Name: "version", Type: "bigint", IsNull: true},
			},
		},
	}
	for _, table := range tables {
		table := table
		t.Run(table.name, func(t *testing.T) {
			var desc []column
			_, err := db.db.Query(&desc,
				`SELECT column_name, data_type, is_nullable::bool , column_default
				   FROM INFORMATION_SCHEMA.COLUMNS
				  WHERE table_name = ?
				  ORDER BY column_name`, table.name,
			)
			require.Nil(t, err)
			require.Equal(t, table.description, desc)
		})
	}
}

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = ""
	dbDatabase = "postgres"
)

const monthID = 1

func initDB(t *testing.T) *DB {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	// Discard log messages in tests
	log.SetOutput(ioutil.Discard)

	config := Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbDatabase,
	}
	db, err := NewDB(config, log)
	require.Nil(t, err)
	err = db.Prepare()
	require.Nil(t, err)

	return db
}

func cleanUp(t *testing.T, db *DB) {
	err := db.DropDB()
	require.Nil(t, err)

	err = db.Shutdown()
	require.Nil(t, err)
}
