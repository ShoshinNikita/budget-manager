// package base provides a db-agnostic implementation of the storage
package base

import (
	"context"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/migrator"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

const (
	Question = sqlx.Question
	Dollar   = sqlx.Dollar
)

type DB struct {
	db *sqlx.DB
}

// NewDB creates a new connection to the db and applies the migrations
func NewDB(driverName, dataSourceName string, placeholder sqlx.Placeholder,
	migrations []*migrator.Migration, log logger.Logger) (*DB, error) {

	conn, err := sqlx.Open(driverName, dataSourceName, placeholder, log)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open a connection to db")
	}

	ctx := context.Background()

	if err := pingDB(ctx, conn, log); err != nil {
		return nil, err
	}
	if err := applyMigrations(ctx, conn, migrations, log); err != nil {
		return nil, err
	}

	return &DB{conn}, nil
}

const (
	pingRetries      = 10
	pingRetryTimeout = 500 * time.Millisecond
)

func pingDB(ctx context.Context, db *sqlx.DB, log logger.Logger) error {
	for i := 0; i < pingRetries; i++ {
		err := db.Ping(ctx)
		if err == nil {
			break
		}

		log.WithError(err).WithField("try", i+1).Debug("couldn't ping DB")
		if i+1 == pingRetries {
			// Don't sleep extra time
			return errors.New("database is down")
		}

		time.Sleep(pingRetryTimeout)
	}
	return nil
}

func applyMigrations(ctx context.Context, db *sqlx.DB, migrations []*migrator.Migration, log logger.Logger) error {
	migrator, err := migrator.NewMigrator(db.GetInternalDB(), log, migrations)
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
