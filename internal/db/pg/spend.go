package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args db_common.AddSpendArgs) (id uint, err error) {
	if !db.checkDay(args.DayID) {
		return 0, db_common.ErrDayNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add Spend
		id, err = db.addSpend(tx, args)
		if err != nil {
			return err
		}

		// Recompute Month budget
		monthID, err := db.getMonthIDByDayID(ctx, tx, args.DayID)
		if err != nil {
			return errors.Wrap(err, "couldn't get Month which contains Day with passed id")
		}

		if err := db.recomputeMonth(tx, monthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (DB) addSpend(tx *pg.Tx, args db_common.AddSpendArgs) (uint, error) {
	spend := &Spend{
		DayID:  args.DayID,
		Title:  args.Title,
		TypeID: args.TypeID,
		Notes:  args.Notes,
		Cost:   args.Cost,
	}
	if err := tx.Insert(spend); err != nil {
		return 0, err
	}
	return spend.ID, nil
}

// EditSpend edits existeng Spend
func (db DB) EditSpend(ctx context.Context, args db_common.EditSpendArgs) error {
	if !db.checkSpend(args.ID) {
		return db_common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		spend := &Spend{ID: args.ID}

		if err := tx.Select(spend); err != nil {
			return errors.Wrap(err, "couldn't select Spend to edit")
		}

		// Edit Spend
		if err := db.editSpend(tx, spend, args); err != nil {
			return err
		}

		// Recompute Month budget
		monthID, err := db.getMonthIDByDayID(ctx, tx, spend.DayID)
		if err != nil {
			return errors.Wrap(err, "couldn't get Month which contains Day with passed dayID")
		}

		if err := db.recomputeMonth(tx, monthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
}

func (DB) editSpend(tx *pg.Tx, spend *Spend, args db_common.EditSpendArgs) error {
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
func (db DB) RemoveSpend(ctx context.Context, id uint) error {
	if !db.checkSpend(id) {
		return db_common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		spend := &Spend{ID: id}

		// Select day id
		err := tx.Model(spend).Column("day_id").WherePK().Select()
		if err != nil {
			return errors.Wrap(err, "couldn't select Spend to remove")
		}

		// Remove Spend
		if err := tx.Delete(spend); err != nil {
			return err
		}

		// Recompute Month budget
		monthID, err := db.getMonthIDByDayID(ctx, tx, spend.DayID)
		if err != nil {
			return errors.Wrap(err, "couldn't get Month which contains Day with passed id")
		}

		if err := db.recomputeMonth(tx, monthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
}
