package db

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	ID uint `pg:",pk"`

	Year  int
	Month time.Month

	Title  string
	Notes  string
	Income int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*Income)(nil)

func (in *Income) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Skip Year

	// Check Month
	if !(time.January <= in.Month && in.Month <= time.December) {
		return ctx, errors.Errorf("invalid month: '%d'", in.Month)
	}

	// Check Title
	if in.Title == "" {
		return ctx, errors.New("title can't be empty")
	}

	// Skip Notes

	// Check Income
	if in.Income <= 0 {
		return ctx, errors.Errorf("invalid income: '%d'", in.Income)
	}

	return ctx, nil
}

// Spend contains information about spends
type Spend struct {
	ID uint `pg:",pk"`

	Year  int
	Month time.Month
	Day   int16

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*Spend)(nil)

func (in *Spend) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Skip Year

	// Check Month
	if !(time.January <= in.Month && in.Month <= time.December) {
		return ctx, errors.Errorf("invalid month: '%d'", in.Month)
	}

	// Check Day
	if !(0 < in.Day && int(in.Day) <= daysInMonth(in.Month)) {
		return ctx, errors.Errorf("invalid day: '%d'", in.Day)
	}

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

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	ID uint `pg:",pk"`

	Year  int
	Month time.Month

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*MonthlyPayment)(nil)

func (in *MonthlyPayment) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Skip Year

	// Check Month
	if !(time.January <= in.Month && in.Month <= time.December) {
		return ctx, errors.Errorf("invalid month: '%d'", in.Month)
	}

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

// SpendType contains information about spend type
type SpendType struct {
	ID   uint `pg:",pk"`
	Name string
}

var _ orm.BeforeInsertHook = (*SpendType)(nil)

func (in *SpendType) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Name
	if in.Name == "" {
		return ctx, errors.Errorf("name can't be empty")
	}

	return ctx, nil
}
