package db

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/db/money"
)

var (
	_ orm.BeforeInsertHook = (*MonthlyPayment)(nil)
	_ orm.BeforeUpdateHook = (*MonthlyPayment)(nil)
)

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint `json:"month_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	TypeID uint        `json:"type_id,omitempty"`
	Type   *SpendType  `pg:"fk:type_id" json:"type,omitempty"`
	Notes  string      `json:"notes,omitempty"`
	Cost   money.Money `json:"cost"`
}

func (in *MonthlyPayment) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Title
	if in.Title == "" {
		return ctx, badRequestError(errors.New("title can't be empty"))
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if in.Cost <= 0 {
		return ctx, badRequestError(errors.Errorf("invalid income: '%d'", in.Cost))
	}

	return ctx, nil
}

func (in *MonthlyPayment) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return in.BeforeInsert(ctx)
}

// -----------------------------------------------------------------------------

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
			err = errorWrap(err, "can't add new MonthlyPayment")
			db.log.Error(err)
			return err
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, args.MonthID)
		if err != nil {
			err = errRecomputeBudget(err)
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

	return id, nil
}

func (_ DB) addMonthlyPayment(tx *pg.Tx, args AddMonthlyPaymentArgs) (id uint, err error) {
	mp := &MonthlyPayment{
		MonthID: args.MonthID,
		Title:   args.Title,
		Notes:   args.Notes,
		TypeID:  args.TypeID,
		Cost:    args.Cost,
	}

	err = tx.Insert(mp)
	if err != nil {
		return 0, errorWrap(err, "can't insert MonthlyPayment")
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

	mp := &MonthlyPayment{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Select(mp)
		if err != nil {
			if err == pg.ErrNoRows {
				err = ErrMonthlyPaymentNotExist
			} else {
				err = errorWrap(err, "can't get Monthly Payment with passed id")
			}

			db.log.Error(err)
			return err
		}

		// Edit Monthly Payment
		err = db.editMonthlyPayment(tx, mp, args)
		if err != nil {
			err = errorWrapf(err, "can't edit Monthly Payment with id '%d'", args.ID)
			db.log.Error(err)
			return err
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, mp.MonthID)
		if err != nil {
			err = errRecomputeBudget(err)
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

func (_ DB) editMonthlyPayment(tx *pg.Tx, mp *MonthlyPayment, args EditMonthlyPaymentArgs) error {
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

	err := tx.Update(mp)
	return errorWrap(err, "can't update Monthly Payment")
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(id uint) error {
	if !db.checkMonthlyPayment(id) {
		return ErrMonthlyPaymentNotExist
	}

	mp := &MonthlyPayment{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Remove Monthly Payment
		err = tx.Model(mp).Column("month_id").WherePK().Select()
		if err != nil {
			err = errorWrap(err, "can't select Monthly Payment with passed id")
			db.log.Error(err)
			return err
		}

		err = db.db.Delete(mp)
		if err != nil {
			err = errorWrapf(err, "can't remove Monthly Payment with id '%d'", id)
			db.log.Error(err)
			return err
		}

		// Recompute month budget
		err = db.recomputeMonth(tx, mp.MonthID)
		if err != nil {
			err = errRecomputeBudget(err)
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
