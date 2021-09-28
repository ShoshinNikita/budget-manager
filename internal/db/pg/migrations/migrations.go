package migrations

import "github.com/lopezator/migrator"

func GetMigrations() []*migrator.Migration {
	return []*migrator.Migration{
		{
			Name: "init",
			Func: initMigration,
		},
		{
			Name: "add NOT NULL constraints",
			Func: addNotNullMigration,
		},
		{
			Name: "add FOREIGN KEY constraints",
			Func: addForeignKeysMigration,
		},
		{
			Name: "add support of nested types",
			Func: addParentIDToSpendTypesMigration,
		},
	}
}
