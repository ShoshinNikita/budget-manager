package db

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

type Month struct {
	ID uint

	Year  int
	Month time.Month

	Incomes         []*Income         `pg:"fk:month_id"`
	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`

	// DailyBudget is a (Sum of Incomes - Cost of Monthly Payments) / Number of Days
	// multiplied by 100
	DailyBudget int64
	Days        []*Day `pg:"fk:month_id"`
}

type Day struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint

	Day int
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  int64
	Spends []*Spend `pg:"fk:day_id"`
}

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	Notes  string
	Income int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*Income)(nil)

func (in *Income) BeforeInsert(ctx context.Context) (context.Context, error) {
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
	// MonthID is a foreign key to Days table
	DayID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*Spend)(nil)

func (in *Spend) BeforeInsert(ctx context.Context) (context.Context, error) {
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
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

var _ orm.BeforeInsertHook = (*MonthlyPayment)(nil)

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
