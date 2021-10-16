// Package sqlx is a wrapper for 'github.com/jmoiron/sqlx' that exports only required methods
package sqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

type Placeholder int

const (
	Question Placeholder = iota
	Dollar
)

type DB struct {
	db          *sqlx.DB
	placeholder Placeholder
	log         logger.Logger
}

func Open(driverName, dataSourceName string, placeholder Placeholder, log logger.Logger) (*DB, error) {
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{db, placeholder, log}, nil
}

func (db DB) Close() error {
	return db.db.Close()
}

// GetInternalDB returns the underlying *sql.DB
func (db DB) GetInternalDB() *sql.DB {
	return db.db.DB
}

func (db DB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

func (db DB) RunInTransaction(ctx context.Context, fn func(*Tx) error) (err error) {
	rollback := func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil {
			db.log.WithError(err).Error("couldn't rollback tx")
		}
	}

	tx, err := db.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "couldn't begin tx")
	}
	defer func() {
		if r := recover(); r != nil {
			rollback(tx)
			panic(r)
		}
	}()

	if err := fn(&Tx{tx: tx, placeholder: db.placeholder}); err != nil {
		rollback(tx)
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "couldn't commit tx")
	}
	return nil
}

type Tx struct {
	tx          *sqlx.Tx
	placeholder Placeholder
}

func (tx Tx) Get(dest interface{}, query string, args ...interface{}) error {
	return tx.GetQuery(dest, newRawQuery(query, args...))
}

func (tx Tx) Select(dest interface{}, query string, args ...interface{}) error {
	return tx.SelectQuery(dest, newRawQuery(query, args...))
}

func (tx Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecQuery(newRawQuery(query, args...))
}

type Sqlizer interface {
	ToSQL() (query string, args []interface{}, err error)
}

func (tx Tx) GetQuery(dest interface{}, query Sqlizer) error {
	sql, args, err := tx.prepareQuery(query)
	if err != nil {
		return err
	}
	return tx.tx.Get(dest, sql, args...)
}

func (tx Tx) SelectQuery(dest interface{}, query Sqlizer) error {
	sql, args, err := tx.prepareQuery(query)
	if err != nil {
		return err
	}
	return tx.tx.Select(dest, sql, args...)
}

func (tx Tx) ExecQuery(query Sqlizer) (sql.Result, error) {
	sql, args, err := tx.prepareQuery(query)
	if err != nil {
		return nil, err
	}
	return tx.tx.Exec(sql, args...)
}

func (tx Tx) prepareQuery(query Sqlizer) (sql string, args []interface{}, err error) {
	sql, args, err = query.ToSQL()
	if err != nil {
		return "", nil, errors.Wrap(err, "ToSql method failed")
	}

	switch tx.placeholder {
	case Question:
		// Questions is used by default
	case Dollar:
		sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	default:
		err = errors.Errorf("unexpected placeholder: %v", tx.placeholder)
	}
	if err != nil {
		return "", nil, errors.Wrap(err, "couldn't replace placeholders")
	}
	return sql, args, err
}

type rawQuery struct {
	query string
	args  []interface{}
}

func (q rawQuery) ToSQL() (string, []interface{}, error) {
	return q.query, q.args, nil
}

func newRawQuery(query string, args ...interface{}) Sqlizer {
	return rawQuery{
		query: query,
		args:  args,
	}
}

type inQuery struct {
	query string
	args  []interface{}
}

func (q inQuery) ToSQL() (string, []interface{}, error) {
	return sqlx.In(q.query, q.args...)
}

func In(query string, args ...interface{}) Sqlizer {
	return inQuery{query: query, args: args}
}
