package db

import (
	"context"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var _ orm.BeforeInsertHook = (*Spend)(nil)

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

// -----------------------------------------------------------------------------

func (db DB) AddSpend() {}

func (db DB) EditSpend() {}

func (db DB) RemoveSpend() {}
