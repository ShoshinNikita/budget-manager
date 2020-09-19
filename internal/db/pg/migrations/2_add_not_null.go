package migrations

import (
	"github.com/go-pg/migrations/v8"
)

func registerAddNotNull(migrator *migrations.Collection) {
	migrator.MustRegisterTx(addNotNullUp, addNotNullDown)
}

func addNotNullUp(db migrations.DB) error {
	_, err := db.Exec(`
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

func addNotNullDown(db migrations.DB) error {
	_, err := db.Exec(`
		ALTER TABLE months
		    ALTER COLUMN year DROP NOT NULL,
		    ALTER COLUMN month DROP NOT NULL,
		    ALTER COLUMN daily_budget DROP NOT NULL,
		    ALTER COLUMN daily_budget DROP DEFAULT 0,
		    ALTER COLUMN total_income DROP NOT NULL,
		    ALTER COLUMN total_income DROP DEFAULT 0,
		    ALTER COLUMN total_spend DROP NOT NULL,
		    ALTER COLUMN total_spend DROP DEFAULT 0,
		    ALTER COLUMN result DROP NOT NULL,
		    ALTER COLUMN result DROP DEFAULT 0;

		ALTER TABLE days
		    ALTER COLUMN month_id DROP NOT NULL,
		    ALTER COLUMN day DROP NOT NULL,
		    ALTER COLUMN saldo DROP NOT NULL;

		ALTER TABLE incomes
		    ALTER COLUMN month_id DROP NOT NULL,
		    ALTER COLUMN title DROP NOT NULL,
		    ALTER COLUMN income DROP NOT NULL;

		ALTER TABLE monthly_payments
		    ALTER COLUMN month_id DROP NOT NULL,
		    ALTER COLUMN title DROP NOT NULL,
		    ALTER COLUMN cost DROP NOT NULL;

		ALTER TABLE spends
		    ALTER COLUMN day_id DROP NOT NULL,
		    ALTER COLUMN title DROP NOT NULL,
		    ALTER COLUMN cost DROP NOT NULL;

		ALTER TABLE spend_types
		    ALTER COLUMN name DROP NOT NULL;`,
	)
	return err
}
