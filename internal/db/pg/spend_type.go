package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(_ context.Context, id uint) (*SpendType, error) {
	log := db.log.WithField("id", id)

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

		log.WithError(err).Error("couldn't get Spend Type")
		return nil, err
	}

	log.Debug("return the Spend Types")
	return spendType, nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(_ context.Context) ([]SpendType, error) {
	log := db.log

	spendTypes := []SpendType{}
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
	if err != nil {
		err = errors.Wrap(err,
			errors.WithMsg("can't select Spend Types"),
			errors.WithType(errors.AppError))

		log.WithError(err).Error("coudln't get all Spend Types")
		return nil, err
	}

	log.Debug("return all Spend Types")
	return spendTypes, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(_ context.Context, name string) (typeID uint, err error) {
	log := db.log.WithField("name", name)

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
		log.WithError(err).Error("coudln't add a new Spend Type")
		return 0, err
	}

	log.WithField("id", typeID).Info("a new Spend Type was successfully created")
	return spendType.ID, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(_ context.Context, id uint, newName string) error {
	log := db.log.WithFields(logrus.Fields{
		"id":       id,
		"new_name": newName,
	})

	if !db.checkSpendType(id) {
		err := ErrSpendTypeNotExist
		log.Error(err)
		return err
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
		log.WithError(err).Error("coudldn't edit the Spend Type")
		return err
	}

	log.Info("the Spend Type was successfully edited")
	return nil
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(_ context.Context, id uint) error {
	log := db.log.WithField("id", id)

	if !db.checkSpendType(id) {
		err := ErrSpendTypeNotExist
		log.Error(err)
		return err
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
		log.WithError(err).Error("coudln't remove the Spend Type")
		return err
	}

	log.Info("the Spend Type was successfully removed")
	return nil
}
