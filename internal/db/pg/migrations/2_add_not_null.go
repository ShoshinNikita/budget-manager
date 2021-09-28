package migrations

import (
	"database/sql"
)

func addNotNullMigration(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE months
		    ALTER COLUMN year SET NOT NULL,
		    ALTER COLUMN month SET NOT NULL,
		    ALTER COLUMN daily_budget SET NOT NULL,
		    ALTER COLUMN daily_budget SET DEFAULT 0,
		    ALTER COLUMN total_income SET NOT NULL,
		    ALTER COLUMN total_income SET DEFAULT 0,
		    ALTER COLUMN total_spend SET NOT NULL,
		    ALTER COLUMN total_spend SET DEFAULT 0,
		    ALTER COLUMN result SET NOT NULL,
		    ALTER COLUMN result SET DEFAULT 0;

		ALTER TABLE days
		    ALTER COLUMN month_id SET NOT NULL,
		    ALTER COLUMN day SET NOT NULL,
		    ALTER COLUMN saldo SET NOT NULL,
		    ALTER COLUMN saldo SET DEFAULT 0;

		ALTER TABLE incomes
		    ALTER COLUMN month_id SET NOT NULL,
		    ALTER COLUMN title SET NOT NULL,
			ALTER COLUMN income SET NOT NULL;

		ALTER TABLE monthly_payments
		    ALTER COLUMN month_id SET NOT NULL,
		    ALTER COLUMN title SET NOT NULL,
		    ALTER COLUMN cost SET NOT NULL;

		ALTER TABLE spends
		    ALTER COLUMN day_id SET NOT NULL,
		    ALTER COLUMN title SET NOT NULL,
		    ALTER COLUMN cost SET NOT NULL;

		ALTER TABLE spend_types
		    ALTER COLUMN name SET NOT NULL;`,
	)
	return err
}
