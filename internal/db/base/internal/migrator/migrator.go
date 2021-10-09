package migrator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lopezator/migrator"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

const (
	// migrationTable is an old migration table used by 'github.com/go-pg/migrations/v8'
	migrationTableV1 = "migrations"

	// migrationTableV2 is a new migration table used by 'github.com/lopezator/migrator'
	migrationTableV2 = "migrations_v2"
)

type Migration = migrator.Migration

type Migrator struct {
	db *sql.DB
	m  *migrator.Migrator
}

// NewMigrator creates a new migrator and registers all migrations
func NewMigrator(db *sql.DB, log logger.Logger, migrations []*migrator.Migration) (*Migrator, error) {
	m, err := migrator.New(
		migrator.TableName(migrationTableV2),
		migrator.WithLogger(migratorLogger{log}),
		migrator.Migrations(convertMigrations(migrations)...),
	)
	if err != nil {
		return nil, err
	}
	return &Migrator{db, m}, nil
}

func convertMigrations(migrations []*migrator.Migration) []interface{} {
	res := make([]interface{}, 0, len(migrations))
	for _, m := range migrations {
		res = append(res, m)
	}
	return res
}

func (m Migrator) Migrate(ctx context.Context) error {
	if err := m.migrateMigrationTableV1(ctx); err != nil {
		return errors.Wrap(err, "couldn't migrate migration table v1")
	}

	return m.m.Migrate(m.db)
}

func (m Migrator) migrateMigrationTableV1(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", migrationTableV1))
	return err
}

type migratorLogger struct {
	log logger.Logger
}

func (l migratorLogger) Printf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}
