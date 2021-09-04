package migrations

import (
	"github.com/lopezator/migrator"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

const MigrationTable = "migrations"

// NewMigrator creates a new migrator and registers all migrations
func NewMigrator(log logger.Logger) (*migrator.Migrator, error) {
	return migrator.New(
		migrator.TableName(MigrationTable),
		migrator.WithLogger(migratorLogger{log}),
		migrator.Migrations(
			&migrator.Migration{
				Name: "init",
				Func: initUp,
			},
			&migrator.Migration{
				Name: "add NOT NULL constraints",
				Func: addNotNullUp,
			},
			&migrator.Migration{
				Name: "add FOREIGN KEY constraints",
				Func: addForeignKeysUp,
			},
			&migrator.Migration{
				Name: "add support of nested types",
				Func: addParentIDToSpendTypesUp,
			},
		),
	)
}

type migratorLogger struct {
	log logger.Logger
}

func (l migratorLogger) Printf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}
