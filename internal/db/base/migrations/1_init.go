package migrations

import (
	"database/sql"
)

func initMigration(tx *sql.Tx) error {
	_, err := tx.Exec(`
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
