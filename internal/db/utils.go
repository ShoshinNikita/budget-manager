package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

// createTables create table according to passed models and options. It returns the first encountered error
//
// Example:
//   createTables(
//     db,
//     &firstModel{}, &orm.CreateTableOptions{...},
//     &secondModel{}, nil,
//   )
//
func createTables(db *pg.DB, modelsAndOpts ...interface{}) error {
	if len(modelsAndOpts)%2 != 0 {
		return errors.New("invalid numbers of arguments")
	}

	for i := 0; i < len(modelsAndOpts); i += 2 {
		model := modelsAndOpts[i]
		opts, ok := modelsAndOpts[i+1].(*orm.CreateTableOptions)
		if !ok {
			if modelsAndOpts[i+1] != nil {
				return errors.Errorf("invalid opts type: '%s'", reflect.TypeOf(modelsAndOpts[i+1]))
			}

			opts = nil
		}

		err := db.CreateTable(model, opts)
		if err != nil {
			return errors.Wrap(err, "can't create table")
		}
	}

	return nil
}

// dropTables drops table according to passed models and options. It returns the first encountered error
//
// Example:
//   dropTable(
//     db,
//     &firstModel{}, &orm.DropTableOptions{...},
//     &secondModel{}, nil,
//   )
//
func dropTables(db *pg.DB, modelsAndOpts ...interface{}) error {
	if len(modelsAndOpts)%2 != 0 {
		return errors.New("invalid numbers of arguments")
	}

	for i := 0; i < len(modelsAndOpts); i += 2 {
		model := modelsAndOpts[i]
		opts, ok := modelsAndOpts[i+1].(*orm.DropTableOptions)
		if !ok {
			if modelsAndOpts[i+1] != nil {
				return errors.Errorf("invalid opts type: '%s'", reflect.TypeOf(modelsAndOpts[i+1]))
			}

			opts = nil
		}

		err := db.DropTable(model, opts)
		if err != nil {
			return errors.Wrap(err, "can't drop table")
		}
	}

	return nil
}

// ping checks the connection to the database
func ping(db *pg.DB) (ok bool) {
	_, err := db.Exec("SELECT 1")
	return err == nil
}

// daysInMonth returns number of days in a month
func daysInMonth(m time.Month) int {
	year := time.Now().Year()

	currentMonth := time.Date(year, m, 1, 0, 0, 0, 0, time.Local)
	// Can just use m+1. time.Date will normalize overflowing month
	nextMonth := time.Date(year, m+1, 1, 0, 0, 0, 0, time.Local)

	days := int64(nextMonth.Sub(currentMonth)) / (int64(time.Hour) * 24)

	return int(days)
}

type cronLogger struct {
	log *clog.Logger
}

// Info logs routine messages about cron's operation.
func (c cronLogger) Info(msg string, keysAndValues ...interface{}) {
	args := buildStringFromKeysAndValues(keysAndValues...)
	c.log.Info(msg, args)
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
