package db

import (
	"github.com/go-pg/pg/v9"

	"github.com/ShoshinNikita/budget_manager/internal/db/models"
	"github.com/ShoshinNikita/budget_manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget_manager/internal/pkg/money"
)

type AddMonthlyPaymentArgs struct {
	MonthID uint
	Title   string
	TypeID  uint
	Notes   string
	Cost    money.Money
}

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(args AddMonthlyPaymentArgs) (id uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, ErrMonthNotExist
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
		db.log.Error(err)
		return 0, err
	}

	return id, nil
}

func (DB) addMonthlyPayment(tx *pg.Tx, args AddMonthlyPaymentArgs) (id uint, err error) {
	mp := &models.MonthlyPayment{
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

type EditMonthlyPaymentArgs struct {
	ID uint

	Title  *string
	TypeID *uint
	Notes  *string
	Cost   *money.Money
}

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(args EditMonthlyPaymentArgs) error {
	if !db.checkMonthlyPayment(args.ID) {
		return ErrMonthlyPaymentNotExist
	}

	mp := &models.MonthlyPayment{ID: args.ID}
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
		db.log.Error(err)
		return err
	}

	return nil
}

func (DB) editMonthlyPayment(tx *pg.Tx, mp *models.MonthlyPayment, args EditMonthlyPaymentArgs) error {
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
func (db DB) RemoveMonthlyPayment(id uint) error {
	if !db.checkMonthlyPayment(id) {
		return ErrMonthlyPaymentNotExist
	}

	mp := &models.MonthlyPayment{ID: id}
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
		db.log.Error(err)
		return err
	}

	return nil
}
