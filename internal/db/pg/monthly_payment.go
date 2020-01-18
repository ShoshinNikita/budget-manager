package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(_ context.Context, args AddMonthlyPaymentArgs) (id uint, err error) {
	log := db.log.WithFields(logrus.Fields{
		"month_id": args.MonthID,
		"title":    args.Title,
		"type_id":  args.TypeID,
		"cost":     args.Cost,
	})

	if !db.checkMonth(args.MonthID) {
		err := ErrMonthNotExist
		log.Error(err)
		return 0, err
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add Monthly Payment
		id, err = db.addMonthlyPayment(tx, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't add a new Monthly Payment"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, args.MonthID)
		if err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't add a new Monthly Payment")
		return 0, err
	}

	log.WithField("monthly_payment_id", id).Info("a new Monthly Payment was successfully created")
	return id, nil
}

func (DB) addMonthlyPayment(tx *pg.Tx, args AddMonthlyPaymentArgs) (id uint, err error) {
	mp := &MonthlyPayment{
		MonthID: args.MonthID,
		Title:   args.Title,
		Notes:   args.Notes,
		TypeID:  args.TypeID,
		Cost:    args.Cost,
	}

	if err := checkModel(mp); err != nil {
		return 0, err
	}
	err = tx.Insert(mp)
	if err != nil {
		return 0, err
	}

	return mp.ID, nil
}

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(_ context.Context, args EditMonthlyPaymentArgs) error {
	log := db.log.WithFields(logrus.Fields{
		"id":          args.ID,
		"new_title":   args.Title,
		"new_type_id": args.TypeID,
		"new_cost":    args.Cost,
	})

	if !db.checkMonthlyPayment(args.ID) {
		err := ErrMonthlyPaymentNotExist
		log.Error(err)
		return err
	}

	mp := &MonthlyPayment{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Select(mp)
		if err != nil {
			if err == pg.ErrNoRows {
				return ErrMonthlyPaymentNotExist
			}
			return errors.Wrap(err,
				errors.WithMsg("can't get Monthly Payment with passed id"),
				errors.WithType(errors.AppError))
		}

		// Edit Monthly Payment
		err = db.editMonthlyPayment(tx, mp, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't edit the Monthly Payment"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, mp.MonthID)
		if err != nil {
			return errRecomputeBudget(err)
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't edit the Monthly Payment")
		return err
	}

	log.Info("the Monthly Payment was successfully edited")
	return nil
}

func (DB) editMonthlyPayment(tx *pg.Tx, mp *MonthlyPayment, args EditMonthlyPaymentArgs) error {
	if args.Title != nil {
		mp.Title = *args.Title
	}
	if args.TypeID != nil {
		mp.TypeID = *args.TypeID
	}
	if args.Notes != nil {
		mp.Notes = *args.Notes
	}
	if args.Cost != nil {
		mp.Cost = *args.Cost
	}

	if err := checkModel(mp); err != nil {
		return err
	}

	return tx.Update(mp)
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(_ context.Context, id uint) error {
	log := db.log.WithField("id", id)

	if !db.checkMonthlyPayment(id) {
		err := ErrMonthlyPaymentNotExist
		log.Error(err)
		return err
	}

	mp := &MonthlyPayment{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Remove Monthly Payment
		err = tx.Model(mp).Column("month_id").WherePK().Select()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't select Monthly Payment with passed id"),
				errors.WithType(errors.AppError))
		}

		err = tx.Delete(mp)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't remove Monthly Payment"),
				errors.WithType(errors.AppError))
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, mp.MonthID)
		if err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't remove the Monthly Payment")
		return err
	}

	log.Info("the Monthly Payment was successfully removed")
	return nil
}
