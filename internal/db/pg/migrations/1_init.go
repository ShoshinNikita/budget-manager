package migrations

import (
	"github.com/go-pg/migrations/v8"
)

func registerInit(migrator *migrations.Collection) {
	migrator.MustRegisterTx(initUp, initDown)
}

func initUp(db migrations.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS months (
			id bigserial PRIMARY KEY,
		
			year  bigint,
			month bigint,
		
			daily_budget bigint,
			total_income bigint,
			total_spend  bigint,
			result       bigint
		);

		CREATE TABLE IF NOT EXISTS days (
			id bigserial PRIMARY KEY,
		
			month_id bigint,
		
			day   bigint,
			saldo bigint
		);

		CREATE TABLE IF NOT EXISTS incomes (
			id bigserial PRIMARY KEY,
		
			month_id bigint,
		
			title  text,
			notes  text,
			income bigint
		);

		CREATE TABLE IF NOT EXISTS monthly_payments (
			id bigserial PRIMARY KEY,
		
			month_id bigint,
		
			title    text,
			type_id  bigint,
			notes    text,
			cost     bigint
		);

		CREATE TABLE IF NOT EXISTS spends (
			id bigserial PRIMARY KEY,
		
			day_id bigint,
		
			title   text,
			type_id bigint,
			notes   text,
			cost    bigint
		);

		CREATE TABLE IF NOT EXISTS spend_types (
			id   bigserial PRIMARY KEY,
			name text
		);`,
	)
	return err
}

func initDown(db migrations.DB) error {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS months;
		DROP TABLE IF EXISTS days;
		DROP TABLE IF EXISTS incomes;
		DROP TABLE IF EXISTS monthly_payments;
		DROP TABLE IF EXISTS spends;
		DROP TABLE IF EXISTS spend_types;`,
	)
	return err
}
