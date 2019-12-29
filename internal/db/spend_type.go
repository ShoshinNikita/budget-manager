package db

import (
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

// SpendType contains information about spend type
type SpendType struct {
	ID   uint   `pg:",pk" json:"id"`
	Name string `json:"name"`
}

// Check checks whether Spend Type is valid (not empty name)
func (in SpendType) Check() error {
	// Check Name
	if in.Name == "" {
		return badRequestError(errors.Errorf("name can't be empty"))
	}

	return nil
}

// -----------------------------------------------------------------------------

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(id uint) (*SpendType, error) {
	spendType := &SpendType{ID: id}
	err := db.db.Select(spendType)
	if err != nil {
		err = errorWrapf(err, "can't select Spend Type with id '%d'", id)
		db.log.Error(err)
		return nil, err
	}

	return spendType, nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes() ([]SpendType, error) {
	spendTypes := []SpendType{}
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
	if err != nil {
		err = errorWrap(err, "can't select Spend Types")
		db.log.Error(err)
		return nil, err
	}

	return spendTypes, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(name string) (typeID uint, err error) {
	spendType := &SpendType{Name: name}
	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := spendType.Check(); err != nil {
			return err
		}
		err = tx.Insert(spendType)
		if err != nil {
			err = errorWrap(err, "can't insert a new Spend Type")
			db.log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		if !IsBadRequestError(err) {
			return 0, internalError(err)
		}
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
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		if err := spendType.Check(); err != nil {
			return err
		}
		err = tx.Update(spendType)
		if err != nil {
			err = errorWrap(err, "can't insert a new Spend Type")
			db.log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		if !IsBadRequestError(err) {
			return internalError(err)
		}
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
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Delete(spendType)
		if err != nil {
			err = errorWrapf(err, "can't delete spend type with id '%d'", id)
			db.log.Error(err)
			return err
		}

		// Reset Type IDs

		_, err = tx.Model((*MonthlyPayment)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			err = errorWrap(err, "can't reset Type IDs of Monthly Payments")
			db.log.Error(err)
			return err
		}

		_, err = tx.Model((*Spend)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			err = errorWrap(err, "can't reset Type IDs of Spends")
			db.log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		if !IsBadRequestError(err) {
			return internalError(err)
		}
		return err
	}

	return nil
}
