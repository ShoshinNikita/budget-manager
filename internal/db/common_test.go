package db

import (
	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/go-pg/pg/v9/orm"
	"github.com/stretchr/testify/require"
)

const (
	dbHost     = "localhost"
	dbPort     = "5432"
	dbUser     = "postgres"
	dbPassword = ""
	dbDatabase = "postgres"
)

const monthID = 1

func initDB(require *require.Assertions) *DB {
	opts := NewDBOptions{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Database: dbDatabase,
	}

	log := clog.NewProdConfig().SetLevel(clog.LevelWarn).Build()
	db, err := NewDB(opts, log)
	require.Nil(err)

	dropDB(db, require)

	err = db.Prepare()
	require.Nil(err)

	return db
}

func dropDB(db *DB, require *require.Assertions) {
	var err error

	opts := &orm.DropTableOptions{IfExists: true}

	err = db.db.DropTable(&Month{}, opts)
	require.Nil(err)
	err = db.db.DropTable(&Income{}, opts)
	require.Nil(err)
	err = db.db.DropTable(&MonthlyPayment{}, opts)
	require.Nil(err)
	err = db.db.DropTable(&Day{}, opts)
	require.Nil(err)
	err = db.db.DropTable(&Spend{}, opts)
	require.Nil(err)
	err = db.db.DropTable(&SpendType{}, opts)
	require.Nil(err)
}
