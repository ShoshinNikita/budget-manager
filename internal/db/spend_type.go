package db

import (
	"github.com/ShoshinNikita/budget_manager/internal/db/models"
	"github.com/ShoshinNikita/budget_manager/internal/pkg/errors"
	"github.com/go-pg/pg/v9"
)

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(id uint) (*models.SpendType, error) {
	spendType := &models.SpendType{ID: id}
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
func (db DB) GetSpendTypes() ([]models.SpendType, error) {
	spendTypes := []models.SpendType{}
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
func (db DB) AddSpendType(name string) (typeID uint, err error) {
	spendType := &models.SpendType{Name: name}
	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := checkModel(spendType); err != nil {
			return err
		}

		err = tx.Insert(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't insert a new Spend Type"),
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
func (db DB) EditSpendType(id uint, newName string) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &models.SpendType{ID: id, Name: newName}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := checkModel(spendType); err != nil {
			return err
		}

		err = tx.Update(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't insert a new Spend Type"),
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
func (db DB) RemoveSpendType(id uint) error {
	if !db.checkSpendType(id) {
		return ErrSpendTypeNotExist
	}

	spendType := &models.SpendType{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Delete(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't delete Spend Type"),
				errors.WithType(errors.AppError))
		}

		// Reset Type IDs

		_, err = tx.Model((*models.MonthlyPayment)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't reset Type IDs of Monthly Payments"),
				errors.WithType(errors.AppError))
		}

		_, err = tx.Model((*models.Spend)(nil)).
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
