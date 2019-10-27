package db

import (
	"context"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var _ orm.BeforeInsertHook = (*SpendType)(nil)

// SpendType contains information about spend type
type SpendType struct {
	ID   uint `pg:",pk"`
	Name string
}

func (in *SpendType) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Name
	if in.Name == "" {
		return ctx, errors.Errorf("name can't be empty")
	}

	return ctx, nil
}

// -----------------------------------------------------------------------------

func (db DB) AddSpendType() {}

func (db DB) EditSpendType() {}

func (db DB) RemoveSpendType() {}
