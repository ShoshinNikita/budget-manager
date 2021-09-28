package migrations

import "github.com/lopezator/migrator"

func GetMigrations() []*migrator.Migration {
	return []*migrator.Migration{
		{
			Name: "init",
			Func: initMigration,
		},
	}
}
