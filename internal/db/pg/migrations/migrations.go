package migrations

import "github.com/go-pg/migrations/v7"

// Number of registered migrations. It can be used to check whether we registered all migrations
const MigrationNumber = 3

// RegisterMigrations registers all migrations
func RegisterMigrations(migrator *migrations.Collection) {
	registerInit(migrator)
	registerAddNotNull(migrator)
	registerAddForeignKeys(migrator)
}
