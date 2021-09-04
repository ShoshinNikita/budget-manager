package migrations

import (
	"database/sql"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

func addForeignKeysMigration(tx *sql.Tx) error {
	// Prepare for foreign key constraints

	// Reset nonexistent spend type ids
	_, err := tx.Exec(`
		UPDATE spends SET type_id = NULL WHERE type_id NOT IN (SELECT id FROM spend_types);
		UPDATE monthly_payments SET type_id = NULL WHERE type_id NOT IN (SELECT id FROM spend_types);
	`)
	if err != nil {
		return errors.Wrap(err, "couldn't reset nonexistent spend type ids")
	}

	// Add foreign key constraints
	_, err = tx.Exec(`
		ALTER TABLE days
		    ADD FOREIGN KEY (month_id) REFERENCES months(id);

		ALTER TABLE incomes
		    ADD FOREIGN KEY (month_id) REFERENCES months(id);

		ALTER TABLE monthly_payments
		    ADD FOREIGN KEY (month_id) REFERENCES months(id),
		    ADD FOREIGN KEY (type_id) REFERENCES spend_types(id);

		ALTER TABLE spends
		    ADD FOREIGN KEY (day_id) REFERENCES days(id),
		    ADD FOREIGN KEY (type_id) REFERENCES spend_types(id);
	`)
	return err
}
