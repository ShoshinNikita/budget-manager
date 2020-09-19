package migrations

import "github.com/go-pg/migrations/v8"

const migrationTable = "migrations"

// NewMigrator creates a new migrator and registers all migrations
func NewMigrator() *migrations.Collection {
	migrator := migrations.NewCollection().SetTableName(migrationTable).DisableSQLAutodiscover(true)
	registerMigrations(migrator)

	return migrator
}

// Number of registered migrations. It can be used to check whether we registered all migrations
const MigrationNumber = 4

// RegisterMigrations registers all migrations
func registerMigrations(migrator *migrations.Collection) {
	registerInit(migrator)
	registerAddNotNull(migrator)
	registerAddForeignKeys(migrator)
	registerAddParentIDToSpendTypes(migrator)
}
