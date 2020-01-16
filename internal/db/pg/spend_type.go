package pg

import (
	"context"

	"github.com/go-pg/pg/v9"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(_ context.Context, id uint) (*SpendType, error) {
	spendType := &SpendType{ID: id}
	err := db.db.Select(spendType)
	if err != nil {
		if err == pg.ErrNoRows {
			err = errors.Wrap(err,
				errors.WithMsg("Spend Type with passed id doesn't exist"),
				errors.WithType(errors.UserError))
		} else {
			err = errors.Wrap(err,
				errors.WithMsg("can't select Spend Type"),
				errors.WithType(errors.AppError))
		}

		db.log.Error(err)
		return nil, err
	}

	return spendType, nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(_ context.Context) ([]SpendType, error) {
	spendTypes := []SpendType{}
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
	if err != nil {
		err = errors.Wrap(err,
			errors.WithMsg("can't select Spend Types"),
			errors.WithType(errors.AppError))
		db.log.Error(err)
		return nil, err
	}

	return spendTypes, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(_ context.Context, name string) (typeID uint, err error) {
	spendType := &SpendType{Name: name}
	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := checkModel(spendType); err != nil {
			return errors.Wrap(err, errors.WithMsg("can't add a new Spend Type"))
		}

		err = tx.Insert(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't add a new Spend Type"),
				errors.WithType(errors.AppError))
		}

		return nil
	})
	if err != nil {
		db.log.Error(err)
		return 0, err
	}

	return spendType.ID, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(_ context.Context, id uint, newName string) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &SpendType{ID: id, Name: newName}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := checkModel(spendType); err != nil {
			return errors.Wrap(err, errors.WithMsg("can't edit the Spend Type"))
		}

		err = tx.Update(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't edit the Spend Type"),
				errors.WithType(errors.AppError))
		}

		return nil
	})
	if err != nil {
		db.log.Error(err)
		return err
	}

	return nil
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(_ context.Context, id uint) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &SpendType{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Delete(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't delete Spend Type"),
				errors.WithType(errors.AppError))
		}

		// Reset Type IDs

		_, err = tx.Model((*MonthlyPayment)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't reset Type IDs of Monthly Payments"),
				errors.WithType(errors.AppError))
		}

		_, err = tx.Model((*Spend)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't reset Type IDs of Spends"),
				errors.WithType(errors.AppError))
		}

		return nil
	})
	if err != nil {
		db.log.Error(err)
		return err
	}

	return nil
}
