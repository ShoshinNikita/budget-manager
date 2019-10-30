package db

import (
	"context"

	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var (
	_ orm.BeforeInsertHook = (*SpendType)(nil)
	_ orm.BeforeUpdateHook = (*SpendType)(nil)
)

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

func (in *SpendType) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return in.BeforeInsert(ctx)
}

// -----------------------------------------------------------------------------

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(name string) (typeID uint, err error) {
	spendType := &SpendType{Name: name}
	err = db.db.Insert(spendType)
	if err != nil {
		err = errors.Wrap(err, "can't insert a new Spend Type")
		db.log.Error(err)
		return 0, err
	}

	return spendType.ID, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(id uint, newName string) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &SpendType{ID: id, Name: newName}
	err := db.db.Update(spendType)
	if err != nil {
		err = errors.Wrap(err, "can't insert a new Spend Type")
		db.log.Error(err)
		return err
	}

	return nil
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(id uint) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &SpendType{ID: id}
	err := db.db.Delete(spendType)
	if err != nil {
		err = errors.Wrapf(err, "can't delete spend type with id '%d'", id)
		db.log.Error(err)
		return err
	}

	return nil
}
