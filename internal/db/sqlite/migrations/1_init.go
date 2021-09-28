package migrations

import "database/sql"

func initMigration(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS months (
			id           INTEGER PRIMARY KEY,
			year         INTEGER NOT NULL,
			month        INTEGER NOT NULL,
			daily_budget INTEGER NOT NULL DEFAULT 0,
			total_income INTEGER NOT NULL DEFAULT 0,
			total_spend  INTEGER NOT NULL DEFAULT 0,
			result       INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS days (
			id       INTEGER PRIMARY KEY,
			month_id INTEGER NOT NULL,
			day      INTEGER NOT NULL,
			saldo    INTEGER NOT NULL DEFAULT 0,

			FOREIGN KEY (month_id) REFERENCES months(id)
		);

		CREATE TABLE IF NOT EXISTS incomes (
			id       INTEGER PRIMARY KEY,
			month_id INTEGER NOT NULL,
			title    TEXT NOT NULL,
			notes    TEXT,
			income   INTEGER NOT NULL,

			FOREIGN KEY (month_id) REFERENCES months(id)
		);

		CREATE TABLE IF NOT EXISTS monthly_payments (
			id       INTEGER PRIMARY KEY,
			month_id INTEGER NOT NULL,
			title    TEXT NOT NULL,
			type_id  INTEGER,
			notes    TEXT,
			cost     INTEGER NOT NULL,

			FOREIGN KEY (month_id) REFERENCES months(id),
			FOREIGN KEY (type_id) REFERENCES spend_types(id)
		);

		CREATE TABLE IF NOT EXISTS spends (
			id      INTEGER PRIMARY KEY,
			day_id  INTEGER NOT NULL,
			title   TEXT NOT NULL,
			type_id INTEGER,
			notes   TEXT,
			cost    INTEGER NOT NULL,

			FOREIGN KEY (day_id) REFERENCES days(id),
			FOREIGN KEY (type_id) REFERENCES spend_types(id)
		);

		CREATE TABLE IF NOT EXISTS spend_types (
			id        INTEGER PRIMARY KEY,
			name      TEXT NOT NULL,
			parent_id INTEGER,

			FOREIGN KEY (parent_id) REFERENCES spend_types(id)
		);`,
	)
	return err
}
