package migrations

import "database/sql"

func addParentIDToSpendTypesUp(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE spend_types ADD COLUMN parent_id bigint REFERENCES spend_types(id);`)
	return err
}
