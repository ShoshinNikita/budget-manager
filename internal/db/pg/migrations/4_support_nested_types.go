package migrations

import "database/sql"

func addParentIDToSpendTypesMigration(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE spend_types ADD COLUMN IF NOT EXISTS parent_id bigint REFERENCES spend_types(id);`)
	return err
}
