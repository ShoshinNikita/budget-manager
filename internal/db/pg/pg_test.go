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
	db := initDB(t)
	defer cleanUp(t, db)

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
				{Name: "daily_budget", Type: "bigint", IsNull: true},
				{Name: "id", Type: "bigint", Default: "nextval('months_id_seq'::regclass)"},
				{Name: "month", Type: "bigint", IsNull: true},
				{Name: "result", Type: "bigint", IsNull: true},
				{Name: "total_income", Type: "bigint", IsNull: true},
				{Name: "total_spend", Type: "bigint", IsNull: true},
				{Name: "year", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "days",
			description: []column{
				{Name: "day", Type: "bigint", IsNull: true},
				{Name: "id", Type: "bigint", Default: "nextval('days_id_seq'::regclass)"},
				{Name: "month_id", Type: "bigint", IsNull: true},
				{Name: "saldo", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "incomes",
			description: []column{
				{Name: "id", Type: "bigint", Default: "nextval('incomes_id_seq'::regclass)"},
				{Name: "income", Type: "bigint", IsNull: true},
				{Name: "month_id", Type: "bigint", IsNull: true},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text", IsNull: true},
			},
		},
		{
			name: "monthly_payments",
			description: []column{
				{Name: "cost", Type: "bigint", IsNull: true},
				{Name: "id", Type: "bigint", Default: "nextval('monthly_payments_id_seq'::regclass)"},
				{Name: "month_id", Type: "bigint", IsNull: true},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text", IsNull: true},
				{Name: "type_id", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "spends",
			description: []column{
				{Name: "cost", Type: "bigint", IsNull: true},
				{Name: "day_id", Type: "bigint", IsNull: true},
				{Name: "id", Type: "bigint", Default: "nextval('spends_id_seq'::regclass)"},
				{Name: "notes", Type: "text", IsNull: true},
				{Name: "title", Type: "text", IsNull: true},
				{Name: "type_id", Type: "bigint", IsNull: true},
			},
		},
		{
			name: "spend_types",
			description: []column{
				{Name: "id", Type: "bigint", Default: "nextval('spend_types_id_seq'::regclass)"},
				{Name: "name", Type: "text", IsNull: true},
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
	err = db.DropDB()
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
