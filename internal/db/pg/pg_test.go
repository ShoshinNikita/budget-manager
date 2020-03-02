// +build integration

package pg

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCreatedTables(t *testing.T) {
	// initDB functions calls method 'DB.Prepare' which already have all needed checks

	t.Run("run migrations against the empty db", func(t *testing.T) {
		db := initDB(t)
		cleanUp(t, db)
	})

	t.Run("run migrations against the up-to-date db", func(t *testing.T) {
		db := initDB(t)
		require.Nil(t, db.Shutdown())

		db = initDB(t)
		cleanUp(t, db)
	})
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
