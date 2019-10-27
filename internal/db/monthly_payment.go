package db

import (
	"context"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var _ orm.BeforeInsertHook = (*MonthlyPayment)(nil)

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

func (in *MonthlyPayment) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Title
	if in.Title == "" {
		return ctx, errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if in.Cost <= 0 {
		return ctx, errors.Errorf("invalid income: '%d'", in.Cost)
	}

	return ctx, nil
}

// -----------------------------------------------------------------------------

func (db DB) AddMonthlyPayment() {}

func (db DB) EditMonthlyPayment() {}

func (db DB) RemoveMonthlyPayment() {}
