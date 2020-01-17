package pg

import (
	"context"

	"github.com/go-pg/pg/v9"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args AddSpendArgs) (id uint, err error) {
	if !db.checkDay(args.DayID) {
		return 0, ErrDayNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add Spend
		id, err = db.addSpend(tx, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't add a new Spend"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute Month budget

		monthID, err := db.getMonthIDByDayID(ctx, tx, args.DayID)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't get Month which contains Day with passed dayID"),
				errors.WithType(errors.AppError))
		}

		err = db.recomputeMonth(tx, monthID)
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

func (DB) addSpend(tx *pg.Tx, args AddSpendArgs) (uint, error) {
	spend := &Spend{
		DayID:  args.DayID,
		Title:  args.Title,
		TypeID: args.TypeID,
		Notes:  args.Notes,
		Cost:   args.Cost,
	}

	if err := checkModel(spend); err != nil {
		return 0, err
	}
	err := tx.Insert(spend)
	if err != nil {
		return 0, err
	}

	return spend.ID, nil
}

// EditSpend edits existeng Spend
func (db DB) EditSpend(ctx context.Context, args EditSpendArgs) error {
	if !db.checkSpend(args.ID) {
		return ErrSpendNotExist
	}

	spend := &Spend{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		err = tx.Select(spend)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't select Spend with passed id"),
				errors.WithType(errors.AppError))
		}

		// Edit Spend
		err = db.editSpend(tx, spend, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't edit the Spend"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute Month budget

		monthID, err := db.getMonthIDByDayID(ctx, tx, spend.DayID)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't get Month which contains Day with passed dayID"),
				errors.WithType(errors.AppError))
		}

		err = db.recomputeMonth(tx, monthID)
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

func (DB) editSpend(tx *pg.Tx, spend *Spend, args EditSpendArgs) error {
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

	if err := checkModel(spend); err != nil {
		return err
	}
	return tx.Update(spend)
}

// RemoveSpend removes Spend with passed id
func (db DB) RemoveSpend(ctx context.Context, id uint) error {
	if !db.checkSpend(id) {
		return ErrSpendNotExist
	}

	spend := &Spend{ID: id}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Select day id
		err = tx.Model(spend).Column("day_id").WherePK().Select()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't select Spend with passed id"),
				errors.WithType(errors.AppError))
		}

		// Remove Spend
		err = tx.Delete(spend)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't delete Spend with passed id"),
				errors.WithType(errors.AppError))
		}

		// Recompute Month budget

		monthID, err := db.getMonthIDByDayID(ctx, tx, spend.DayID)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't get Month which contains Day with passed dayID"),
				errors.WithType(errors.AppError))
		}

		err = db.recomputeMonth(tx, monthID)
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
