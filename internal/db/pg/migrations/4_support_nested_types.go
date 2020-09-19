package migrations

import (
	"github.com/go-pg/migrations/v8"
)

func registerAddParentIDToSpendTypes(migrator *migrations.Collection) {
	migrator.MustRegisterTx(addParentIDToSpendTypesUp, addParentIDToSpendTypesDown)
}

func addParentIDToSpendTypesUp(db migrations.DB) error {
	_, err := db.Exec(`ALTER TABLE spend_types ADD COLUMN parent_id bigint REFERENCES spend_types(id);`)
	return err
}

func addParentIDToSpendTypesDown(db migrations.DB) error {
	_, err := db.Exec(`ALTER TABLE spend_types DROP COLUMN parent_id;`)
	return err
}
