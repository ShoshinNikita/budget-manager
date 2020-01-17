// +build integration

package pg

import (
	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/stretchr/testify/require"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = ""
	dbDatabase = "postgres"
)

const monthID = 1

func initDB(require *require.Assertions) *DB {
	log := clog.NewProdConfig().SetLevel(clog.LevelWarn).Build()

	config := Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbDatabase,
	}
	db, err := NewDB(config, log)
	require.Nil(err)
	err = db.DropDB()
	require.Nil(err)
	err = db.Prepare()
	require.Nil(err)

	return db
}

func cleanUp(require *require.Assertions, db *DB) {
	err := db.DropDB()
	require.Nil(err)

	err = db.Shutdown()
	require.Nil(err)
}
