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
		const tableNumber = 6

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
		Name     string `pg:"column_name"`
		DataType string `pg:"data_type"`
	}
	tables := []struct {
		name        string
		description []column
	}{
		// Columns must be sorted by name
		{
			name: "months",
			description: []column{
				{"daily_budget", "bigint"},
				{"id", "bigint"},
				{"month", "bigint"},
				{"result", "bigint"},
				{"total_income", "bigint"},
				{"total_spend", "bigint"},
				{"year", "bigint"},
			},
		},
		{
			name: "days",
			description: []column{
				{"day", "bigint"},
				{"id", "bigint"},
				{"month_id", "bigint"},
				{"saldo", "bigint"},
			},
		},
		{
			name: "incomes",
			description: []column{
				{"id", "bigint"},
				{"income", "bigint"},
				{"month_id", "bigint"},
				{"notes", "text"},
				{"title", "text"},
			},
		},
		{
			name: "monthly_payments",
			description: []column{
				{"cost", "bigint"},
				{"id", "bigint"},
				{"month_id", "bigint"},
				{"notes", "text"},
				{"title", "text"},
				{"type_id", "bigint"},
			},
		},
		{
			name: "spends",
			description: []column{
				{"cost", "bigint"},
				{"day_id", "bigint"},
				{"id", "bigint"},
				{"notes", "text"},
				{"title", "text"},
				{"type_id", "bigint"},
			},
		},
		{
			name: "spend_types",
			description: []column{
				{"id", "bigint"},
				{"name", "text"},
			},
		},
	}
	for _, table := range tables {
		table := table
		t.Run(table.name, func(t *testing.T) {
			var desc []column
			_, err := db.db.Query(&desc,
				`SELECT column_name, data_type
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
