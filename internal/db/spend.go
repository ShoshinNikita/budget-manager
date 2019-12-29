package db

import (
	"github.com/go-pg/pg/v9"

	"github.com/ShoshinNikita/budget_manager/internal/db/models"
	"github.com/ShoshinNikita/budget_manager/internal/pkg/money"
)

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

func (DB) addSpend(tx *pg.Tx, args AddSpendArgs) (uint, error) {
	spend := &models.Spend{
		DayID:  args.DayID,
		Title:  args.Title,
		TypeID: args.TypeID,
		Notes:  args.Notes,
		Cost:   args.Cost,
	}

	if err := spend.Check(); err != nil {
		return 0, badRequestError(err)
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

	spend := &models.Spend{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Select(spend)
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

func (DB) editSpend(tx *pg.Tx, spend *models.Spend, args EditSpendArgs) error {
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

	if err := spend.Check(); err != nil {
		return badRequestError(err)
	}
	return tx.Update(spend)
}

// RemoveSpend removes Spend with passed id
func (db DB) RemoveSpend(id uint) error {
	if !db.checkSpend(id) {
		return ErrSpendNotExist
	}

	spend := &models.Spend{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Remove Spend
		err = tx.Model(spend).Column("day_id").WherePK().Select()
		if err != nil {
			return errorWrap(err, "can't select Spend with passed id")
		}

		err = tx.Delete(spend)
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
