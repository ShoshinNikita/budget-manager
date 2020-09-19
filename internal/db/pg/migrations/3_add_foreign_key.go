package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/pkg/errors"
)

func registerAddForeignKeys(migrator *migrations.Collection) {
	migrator.MustRegisterTx(addForeignKeysUp, addForeignKeysDown)
}

func addForeignKeysUp(db migrations.DB) error {
	// Prepare for foreign key constraints

	// Reset nonexistent spend type ids
	_, err := db.Exec(`
		UPDATE spends SET type_id = NULL WHERE type_id NOT IN (SELECT id FROM spend_types);
		UPDATE monthly_payments SET type_id = NULL WHERE type_id NOT IN (SELECT id FROM spend_types);
	`)
	if err != nil {
		return errors.Wrap(err, "couldn't reset nonexistent spend type ids")
	}

	// Add foreign key constraints
	_, err = db.Exec(`
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

func addForeignKeysDown(db migrations.DB) error {
	_, err := db.Exec(`
		ALTER TABLE days             DROP CONSTRAINT IF EXISTS days_month_id_fkey;
		ALTER TABLE incomes          DROP CONSTRAINT IF EXISTS incomes_month_id_fkey;
		ALTER TABLE monthly_payments DROP CONSTRAINT IF EXISTS monthly_payments_month_id_fkey;
		ALTER TABLE spends           DROP CONSTRAINT IF EXISTS spends_day_id_fkey;
		ALTER TABLE spends           DROP CONSTRAINT IF EXISTS spends_type_id_fkey;
	`)
	return err
}
