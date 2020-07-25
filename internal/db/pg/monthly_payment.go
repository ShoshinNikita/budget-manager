package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(_ context.Context, args db_common.AddMonthlyPaymentArgs) (id uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, db_common.ErrMonthNotExist
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
		return 0, err
	}

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
func (db DB) EditMonthlyPayment(_ context.Context, args db_common.EditMonthlyPaymentArgs) error {
	if !db.checkMonthlyPayment(args.ID) {
		return db_common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		mp := &MonthlyPayment{ID: args.ID}

		if err := tx.Select(mp); err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return db_common.ErrMonthlyPaymentNotExist
			}
			return errors.Wrap(err, "couldn't select Monthly Payment to edit")
		}

		// Edit Monthly Payment
		if err := db.editMonthlyPayment(tx, mp, args); err != nil {
			return err
		}

		// Recompute month budget
		if err := db.recomputeMonth(tx, mp.MonthID); err != nil {
			return errRecomputeBudget(err)
		}
		return nil
	})
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
func (db DB) RemoveMonthlyPayment(_ context.Context, id uint) error {
	if !db.checkMonthlyPayment(id) {
		return db_common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		mp := &MonthlyPayment{ID: id}

		// Remove Monthly Payment
		err := tx.Model(mp).Column("month_id").WherePK().Select()
		if err != nil {
			// Monthly Payment must exist
			return errors.Wrap(err, "couldn't select Monthly Payment to remove")
		}

		if err := tx.Delete(mp); err != nil {
			return err
		}

		// Recompute month budget
		if err := db.recomputeMonth(tx, mp.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
}
