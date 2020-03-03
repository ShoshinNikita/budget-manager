package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(ctx context.Context, id uint) (*db_common.SpendType, error) {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithField("id", id)

	spendType := &SpendType{ID: id}
	err := db.db.Select(spendType)
	if err != nil {
		if err == pg.ErrNoRows {
			err = db_common.ErrSpendTypeNotExist
		} else {
			err = errors.Wrap(err,
				errors.WithMsg("couldn't select Spend Type"),
				errors.WithType(errors.AppError))
		}

		log.WithError(errors.GetOriginalError(err)).Error("couldn't get Spend Type")
		return nil, err
	}

	log.Debug("return the Spend Types")
	return spendType.ToCommon(), nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(ctx context.Context) ([]*db_common.SpendType, error) {
	log := request_id.FromContextToLogger(ctx, db.log)

	spendTypes := []SpendType{}
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
	if err != nil {
		err = errors.Wrap(err,
			errors.WithMsg("couldn't select Spend Types"),
			errors.WithType(errors.AppError))

		log.WithError(errors.GetOriginalError(err)).Error("couldn't get all Spend Types")
		return nil, err
	}

	log.Debug("return all Spend Types")
	res := make([]*db_common.SpendType, 0, len(spendTypes))
	for i := range spendTypes {
		res = append(res, spendTypes[i].ToCommon())
	}
	return res, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(ctx context.Context, name string) (typeID uint, err error) {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithField("name", name)

	spendType := &SpendType{Name: name}
	err = db.db.RunInTransaction(func(tx *pg.Tx) error {
		if err := tx.Insert(spendType); err != nil {
			return errors.Wrap(err,
				errors.WithMsg("couldn't add a new Spend Type"),
				errors.WithType(errors.AppError))
		}
		return nil
	})
	if err != nil {
		log.WithError(errors.GetOriginalError(err)).Error("couldn't add a new Spend Type")
		return 0, err
	}

	log.WithField("id", typeID).Info("a new Spend Type was successfully created")
	return spendType.ID, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(ctx context.Context, id uint, newName string) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{"id": id, "new_name": newName})

	if !db.checkSpendType(id) {
		err := db_common.ErrSpendTypeNotExist
		log.Error(err)
		return err
	}

	spendType := &SpendType{ID: id, Name: newName}
	err := db.db.RunInTransaction(func(tx *pg.Tx) error {
		if err := tx.Update(spendType); err != nil {
			return errors.Wrap(err,
				errors.WithMsg("couldn't edit the Spend Type"),
				errors.WithType(errors.AppError))
		}
		return nil
	})
	if err != nil {
		log.WithError(errors.GetOriginalError(err)).Error("coudldn't edit the Spend Type")
		return err
	}

	log.Info("the Spend Type was successfully edited")
	return nil
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(ctx context.Context, id uint) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithField("id", id)

	if !db.checkSpendType(id) {
		err := db_common.ErrSpendTypeNotExist
		log.Error(err)
		return err
	}

	spendType := &SpendType{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Delete(spendType)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("couldn't delete Spend Type"),
				errors.WithType(errors.AppError))
		}

		// Reset Type IDs

		_, err = tx.Model((*MonthlyPayment)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("couldn't reset Type IDs of Monthly Payments"),
				errors.WithType(errors.AppError))
		}

		_, err = tx.Model((*Spend)(nil)).
			Set("type_id = 0").
			Where("type_id = ?", id).
			Update()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("couldn't reset Type IDs of Spends"),
				errors.WithType(errors.AppError))
		}

		return nil
	})
	if err != nil {
		log.WithError(errors.GetOriginalError(err)).Error("couldn't remove the Spend Type")
		return err
	}

	log.Info("the Spend Type was successfully removed")
	return nil
}
