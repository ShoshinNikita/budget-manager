package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(ctx context.Context, args db_common.AddMonthlyPaymentArgs) (id uint, err error) {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{
		"month_id": args.MonthID, "title": args.Title, "type_id": args.TypeID, "cost": args.Cost,
	})

	if !db.checkMonth(args.MonthID) {
		err := db_common.ErrMonthNotExist
		log.Error(err)
		return 0, err
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add Monthly Payment
		id, err = db.addMonthlyPayment(tx, args)
		if err != nil {
			return err
		}

		// Recompute month budget
		if err = db.recomputeMonth(tx, args.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't add a new Monthly Payment")
		return 0, err
	}

	log.WithField("monthly_payment_id", id).Debug("a new Monthly Payment was successfully created")
	return id, nil
}

func (DB) addMonthlyPayment(tx *pg.Tx, args db_common.AddMonthlyPaymentArgs) (id uint, err error) {
	mp := &MonthlyPayment{
		MonthID: args.MonthID,
		Title:   args.Title,
		Notes:   args.Notes,
		TypeID:  args.TypeID,
		Cost:    args.Cost,
	}
	if err := tx.Insert(mp); err != nil {
		return 0, err
	}
	return mp.ID, nil
}

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(ctx context.Context, args db_common.EditMonthlyPaymentArgs) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{
		"id": args.ID, "new_title": args.Title, "new_type_id": args.TypeID, "new_cost": args.Cost,
	})

	if !db.checkMonthlyPayment(args.ID) {
		err := db_common.ErrMonthlyPaymentNotExist
		log.Error(err)
		return err
	}

	mp := &MonthlyPayment{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Select(mp)
		if err != nil {
			if err == pg.ErrNoRows {
				return db_common.ErrMonthlyPaymentNotExist
			}
			return errors.Wrap(err, "couldn't get Monthly Payment with passed id")
		}

		// Edit Monthly Payment
		if err = db.editMonthlyPayment(tx, mp, args); err != nil {
			return err
		}

		// Recompute month budget
		if err = db.recomputeMonth(tx, mp.MonthID); err != nil {
			return errRecomputeBudget(err)
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't edit the Monthly Payment")
		return err
	}

	log.Debug("the Monthly Payment was successfully edited")
	return nil
}

func (DB) editMonthlyPayment(tx *pg.Tx, mp *MonthlyPayment, args db_common.EditMonthlyPaymentArgs) error {
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
	return tx.Update(mp)
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(ctx context.Context, id uint) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithField("id", id)

	if !db.checkMonthlyPayment(id) {
		err := db_common.ErrMonthlyPaymentNotExist
		log.Error(err)
		return err
	}

	mp := &MonthlyPayment{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Remove Monthly Payment
		err = tx.Model(mp).Column("month_id").WherePK().Select()
		if err != nil {
			// Monthly Payment must exist
			return errors.Wrap(err, "couldn't select Monthly Payment with passed id")
		}

		if err = tx.Delete(mp); err != nil {
			return err
		}

		// Recompute month budget
		if err = db.recomputeMonth(tx, mp.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't remove the Monthly Payment")
		return err
	}

	log.Debug("the Monthly Payment was successfully removed")
	return nil
}
