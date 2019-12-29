package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	clog "github.com/ShoshinNikita/go-clog/v3"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	pkgErrors "github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/pkg/errors"
)

// --------------------------------------------------
// DB
// --------------------------------------------------

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
		return pkgErrors.New("invalid numbers of arguments")
	}

	for i := 0; i < len(modelsAndOpts); i += 2 {
		model := modelsAndOpts[i]
		opts, ok := modelsAndOpts[i+1].(*orm.CreateTableOptions)
		if !ok {
			if modelsAndOpts[i+1] != nil {
				return pkgErrors.Errorf("invalid opts type: '%s'", reflect.TypeOf(modelsAndOpts[i+1]))
			}

			opts = nil
		}

		err := db.CreateTable(model, opts)
		if err != nil {
			return pkgErrors.Wrap(err, "can't create table")
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
		return pkgErrors.New("invalid numbers of arguments")
	}

	for i := 0; i < len(modelsAndOpts); i += 2 {
		model := modelsAndOpts[i]
		opts, ok := modelsAndOpts[i+1].(*orm.DropTableOptions)
		if !ok {
			if modelsAndOpts[i+1] != nil {
				return pkgErrors.Errorf("invalid opts type: '%s'", reflect.TypeOf(modelsAndOpts[i+1]))
			}

			opts = nil
		}

		err := db.DropTable(model, opts)
		if err != nil {
			return pkgErrors.Wrap(err, "can't drop table")
		}
	}

	return nil
}

// ping checks the connection to the database
func ping(db *pg.DB) (ok bool) {
	_, err := db.Exec("SELECT 1")
	return err == nil
}

// --------------------------------------------------
// Time
// --------------------------------------------------

// daysInMonth returns number of days in a month
func daysInMonth(t time.Time) int {
	year := t.Year()
	month := t.Month()

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

// --------------------------------------------------
// Checker
// --------------------------------------------------

// checkModel calls 'Check' method of passed model. If error is not nil,// it wraps error
// with 'errors.WithOriginalError()' and 'errors.WithType(errors.UserError)' options
func checkModel(model interface{ Check() error }) error {
	err := model.Check()
	if err != nil {
		return errors.Wrap(err, errors.WithOriginalError(), errors.WithType(errors.UserError))
	}
	return nil
}
