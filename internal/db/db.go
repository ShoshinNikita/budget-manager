package db

import (
	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var (
	ErrDataBaseIsDown = errors.New("database is down")
)

type DB struct {
	db  *pg.DB
	log *clog.Logger
}

// NewDBOptions contains options for NewDB function
type NewDBOptions struct {
	Host string
	Port string

	User     string
	Password string
	Database string
}

// NewDB creates a new connection to the db and pings it
func NewDB(opts NewDBOptions, log *clog.Logger) (*DB, error) {
	db := pg.Connect(&pg.Options{
		Addr:     opts.Host + ":" + opts.Port,
		User:     opts.User,
		Password: opts.Password,
		Database: opts.Database,
	})

	if !ping(db) {
		return nil, ErrDataBaseIsDown
	}

	log = log.WithPrefix("database")

	return &DB{db: db, log: log}, nil
}

// Prepare prepares the database:
//   - create tables
//
func (db *DB) Prepare() error {
	db.log.Debug("create tables")

	err := createTables(
		db.db,
		&Income{}, &orm.CreateTableOptions{IfNotExists: true},
		&Spend{}, &orm.CreateTableOptions{IfNotExists: true},
		&MonthlyPayment{}, &orm.CreateTableOptions{IfNotExists: true},
		&SpendType{}, &orm.CreateTableOptions{IfNotExists: true},
	)

	err = errors.Wrap(err, "couldn't create tables")

	if err != nil {
		db.log.Error(err)
	}

	return err
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	db.log.Debug("close database connection")

	return db.db.Close()
}
