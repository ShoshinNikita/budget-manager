package db

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/db/money"
)

var (
	_ orm.BeforeInsertHook = (*Spend)(nil)
	_ orm.BeforeUpdateHook = (*Spend)(nil)
)

// Spend contains information about spends
type Spend struct {
	// MonthID is a foreign key to Days table
	DayID uint `json:"day_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	TypeID uint        `json:"type_id,omitempty"`
	Type   *SpendType  `pg:"fk:type_id" json:"type,omitempty"`
	Notes  string      `json:"notes,omitempty"`
	Cost   money.Money `json:"cost"`
}

func (in *Spend) BeforeInsert(ctx context.Context) (context.Context, error) {
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

func (in *Spend) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return in.BeforeInsert(ctx)
}

// -----------------------------------------------------------------------------

type AddSpendArgs struct {
	DayID  uint
	Title  string
	TypeID uint   // optional
	Notes  string // optional
	Cost   money.Money
}

// AddSpend adds a new Spend
func (db DB) AddSpend(args AddSpendArgs) (id uint, err error) {
	if !db.checkDay(args.DayID) {
		return 0, ErrDayNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add Spend
		id, err = db.addSpend(tx, args)
		if err != nil {
			err = errorWrap(err, "can't add a new Spend")
			db.log.Error(err)
			return err
		}

		// Recompute Month budget

		monthID, err := db.GetMonthIDByDayID(args.DayID)
		if err != nil {
			return errorWrap(err, "can't get Month which contains Day with passed dayID")
		}

		err = db.recomputeMonth(tx, monthID)
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

func (_ DB) addSpend(tx *pg.Tx, args AddSpendArgs) (uint, error) {
	spend := &Spend{
		DayID:  args.DayID,
		Title:  args.Title,
		TypeID: args.TypeID,
		Notes:  args.Notes,
		Cost:   args.Cost,
	}

	err := tx.Insert(spend)
	if err != nil {
		return 0, errorWrap(err, "can't insert Spend")
	}

	return spend.ID, nil
}

type EditSpendArgs struct {
	ID     uint
	Title  *string
	TypeID *uint
	Notes  *string
	Cost   *money.Money
}

// EditSpend edits existeng Spend
func (db DB) EditSpend(args EditSpendArgs) error {
	if !db.checkSpend(args.ID) {
		return ErrSpendNotExist
	}

	spend := &Spend{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = db.db.Select(spend)
		if err != nil {
			return errorWrap(err, "can't select Spend with passed id")
		}

		// Edit Spend
		err = db.editSpend(tx, spend, args)
		if err != nil {
			err = errorWrap(err, "can't edit Spend with passed id")
			db.log.Error(err)
			return err
		}

		// Recompute Month budget

		monthID, err := db.GetMonthIDByDayID(spend.DayID)
		if err != nil {
			return errorWrap(err, "can't get Month which contains Day with passed dayID")
		}

		err = db.recomputeMonth(tx, monthID)
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

func (_ DB) editSpend(tx *pg.Tx, spend *Spend, args EditSpendArgs) error {
	if args.Title != nil {
		spend.Title = *args.Title
	}
	if args.TypeID != nil {
		spend.TypeID = *args.TypeID
	}
	if args.Notes != nil {
		spend.Notes = *args.Notes
	}
	if args.Cost != nil {
		spend.Cost = *args.Cost
	}

	return tx.Update(spend)
}

// RemoveSpend removes Spend with passed id
func (db DB) RemoveSpend(id uint) error {
	if !db.checkSpend(id) {
		return ErrSpendNotExist
	}

	spend := &Spend{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Remove Spend
		err = db.db.Model(spend).Column("day_id").WherePK().Select()
		if err != nil {
			return errorWrap(err, "can't select Spend with passed id")
		}

		err = db.db.Delete(spend)
		if err != nil {
			return errorWrap(err, "can't delete Spend with passed id")
		}

		// Recompute Month budget

		monthID, err := db.GetMonthIDByDayID(spend.DayID)
		if err != nil {
			return errorWrap(err, "can't get Month which contains Day with passed dayID")
		}

		err = db.recomputeMonth(tx, monthID)
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
