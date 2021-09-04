package pg

import (
	"fmt"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg/internal/sqlx"
)

// --------------------------------------------------
// DB
// --------------------------------------------------

// checkMonth checks if a Month with passed id exists
func checkMonth(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "months", id)
}

// checkDay checks if a Day with passed id exists
func checkDay(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "days", id)
}

// checkIncome checks if an Income with passed id exists
func checkIncome(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "incomes", id)
}

// checkMonthlyPayment checks if a Monthly Payment with passed id exists
func checkMonthlyPayment(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "monthly_payments", id)
}

// checkSpend checks if a Spend with passed id exists
func checkSpend(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "spends", id)
}

// checkSpendType checks if a Spend Type with passed id exists
func checkSpendType(tx *sqlx.Tx, id uint) bool {
	return checkModel(tx, "spend_types", id)
}

// checkModel checks if a model with passed id exists
func checkModel(tx *sqlx.Tx, table string, id uint) bool {
	var c int
	err := tx.Get(&c, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE id = ?`, table), id)
	if err != nil || c == 0 {
		return false
	}
	return true
}

// --------------------------------------------------
// Time
// --------------------------------------------------

// daysInMonth returns number of days in a month
func daysInMonth(year int, month time.Month) int {
	currentMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	// Can just use m+1. time.Date will normalize overflowing month
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local)

	days := int64(nextMonth.Sub(currentMonth)) / (int64(time.Hour) * 24)

	return int(days)
}
