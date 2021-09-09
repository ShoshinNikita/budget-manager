// package base provides a PostgreSQL implementation for DB
package base

import (
	"context"
	"time"

	"github.com/lopezator/migrator"

	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
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
	db         *sqlx.DB
	log        logger.Logger
	migrations []*migrator.Migration
}

// NewDB creates a new connection to the db and pings it
func NewDB(driverName string, dataSourceName string, placeholder sqlx.Placeholder,
	migrations []*migrator.Migration, log logger.Logger) (*DB, error) {

	conn, err := sqlx.Open(driverName, dataSourceName, placeholder, log)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open a connection to db")
	}
	db := &DB{
		log:        log.WithField("db_type", "pg"),
		db:         conn,
		migrations: migrations,
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
	ctx := context.Background()

	migrator, err := migrations.NewMigrator(db.db.GetInternalDB(), db.log, db.migrations)
	if err != nil {
		return errors.Wrap(err, "couldn't prepare migrator")
	}

	if err := migrator.Migrate(ctx); err != nil {
		return errors.Wrap(err, "couldn't apply migrations")
	}

	return nil
}

// Shutdown closes the connection to the db
func (db *DB) Shutdown() error {
	return db.db.Close()
}
