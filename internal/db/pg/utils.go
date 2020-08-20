package pg

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"
)

// --------------------------------------------------
// DB
// --------------------------------------------------

// ping checks the connection to the database
func ping(db *pg.DB) (ok bool) {
	_, err := db.Exec("SELECT 1")
	return err == nil
}

// checkMonth checks if a Month with passed id exists
func (db DB) checkMonth(id uint) (ok bool) {
	return db.checkModel((*Month)(nil), id)
}

// checkDay checks if a Day with passed id exists
func (db DB) checkDay(id uint) (ok bool) {
	return db.checkModel((*Day)(nil), id)
}

// checkIncome checks if an Income with passed id exists
func (db DB) checkIncome(id uint) (ok bool) {
	return db.checkModel((*Income)(nil), id)
}

// checkMonthlyPayment checks if a Monthly Payment with passed id exists
func (db DB) checkMonthlyPayment(id uint) (ok bool) {
	return db.checkModel((*MonthlyPayment)(nil), id)
}

// checkSpend checks if a Spend with passed id exists
func (db DB) checkSpend(id uint) (ok bool) {
	return db.checkModel((*Spend)(nil), id)
}

// checkSpendType checks if a Spend Type with passed id exists
func (db DB) checkSpendType(id uint) (ok bool) {
	return db.checkModel((*SpendType)(nil), id)
}

// checkModel checks if a model with passed id exists
func (db DB) checkModel(model interface{}, id uint) (ok bool) {
	c, err := db.db.Model(model).Where("id = ?", id).Count()
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

// --------------------------------------------------
// Cron Logger
// --------------------------------------------------

type cronLogger struct {
	log logrus.FieldLogger
}

// Info logs routine messages about cron's operation.
func (c cronLogger) Info(msg string, keysAndValues ...interface{}) {
	args := buildStringFromKeysAndValues(keysAndValues...)
	c.log.Debug(msg, args)
}

// Error logs an error condition.
func (c cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	args := buildStringFromKeysAndValues(keysAndValues...)
	c.log.Error(msg, args)
}

func buildStringFromKeysAndValues(keysAndValues ...interface{}) string {
	if len(keysAndValues) == 0 {
		return ""
	}

	b := strings.Builder{}

	for i := 0; i < len(keysAndValues); {
		switch key := keysAndValues[i].(type) {
		case string:
			b.WriteString(key)
		default:
			b.WriteString(fmt.Sprintf("%+v", key))
		}
		b.WriteByte(':')
		b.WriteByte(' ')

		switch value := keysAndValues[i+1].(type) {
		case string:
			b.WriteString(value)
		default:
			b.WriteString(fmt.Sprintf("%+v", value))
		}
		i += 2

		if i < len(keysAndValues) {
			b.WriteByte(';')
			b.WriteByte(' ')
		}
	}

	return b.String()
}
