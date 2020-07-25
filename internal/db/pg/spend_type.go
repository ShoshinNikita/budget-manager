package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(_ context.Context, id uint) (*db_common.SpendType, error) {
	spendType := &SpendType{ID: id}
	if err := db.db.Select(spendType); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			err = db_common.ErrSpendTypeNotExist
		}
		return nil, err
	}

	return spendType.ToCommon(), nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(_ context.Context) ([]*db_common.SpendType, error) {
	spendTypes := []SpendType{}
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
	if err != nil {
		return nil, err
	}

	res := make([]*db_common.SpendType, 0, len(spendTypes))
	for i := range spendTypes {
		res = append(res, spendTypes[i].ToCommon())
	}
	return res, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(_ context.Context, name string) (typeID uint, err error) {
	spendType := &SpendType{Name: name}
	err = db.db.RunInTransaction(func(tx *pg.Tx) error {
		return tx.Insert(spendType)
	})
	if err != nil {
		return 0, err
	}

	return spendType.ID, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(_ context.Context, id uint, newName string) error {
	if !db.checkSpendType(id) {
		return db_common.ErrSpendTypeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		spendType := &SpendType{ID: id, Name: newName}
		return tx.Update(spendType)
	})
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(_ context.Context, id uint) error {
	if !db.checkSpendType(id) {
		return db_common.ErrSpendTypeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		spendType := &SpendType{ID: id}

		if err = tx.Delete(spendType); err != nil {
			return err
		}

		// Reset Type IDs

		_, err = tx.Model((*MonthlyPayment)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err, "couldn't reset Type IDs of Monthly Payments")
		}

		_, err = tx.Model((*Spend)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err, "couldn't reset Type IDs of Spends")
		}

		return nil
	})
}
