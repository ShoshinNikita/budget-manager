package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// AddIncome adds a new income with passed params
func (db DB) AddIncome(_ context.Context, args db_common.AddIncomeArgs) (incomeID uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, db_common.ErrMonthNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add a new Income
		incomeID, err = db.addIncome(tx, args)
		if err != nil {
			return err
		}

		// Recompute Month budget
		if err = db.recomputeMonth(tx, args.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return incomeID, nil
}

func (DB) addIncome(tx *pg.Tx, args db_common.AddIncomeArgs) (incomeID uint, err error) {
	in := &Income{
		MonthID: args.MonthID,

		Title:  args.Title,
		Notes:  args.Notes,
		Income: args.Income,
	}
	if err = tx.Insert(in); err != nil {
		return 0, err
	}
	return in.ID, nil
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(_ context.Context, args db_common.EditIncomeArgs) error {
	if !db.checkIncome(args.ID) {
		return db_common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		in := &Income{ID: args.ID}

		// Select Income
		if err := tx.Select(in); err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return db_common.ErrIncomeNotExist
			}
			return errors.Wrap(err, "couldn't select Income to edit")
		}

		// Edit Income
		if err := db.editIncome(tx, in, args); err != nil {
			return err
		}

		// Recompute Month budget
		if err := db.recomputeMonth(tx, in.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
}

func (DB) editIncome(tx *pg.Tx, in *Income, args db_common.EditIncomeArgs) error {
	if args.Title != nil {
		in.Title = *args.Title
	}
	if args.Notes != nil {
		in.Notes = *args.Notes
	}
	if args.Income != nil {
		in.Income = *args.Income
	}

	if err := tx.Update(in); err != nil {
		return err
	}
	return nil
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(_ context.Context, id uint) error {
	if !db.checkIncome(id) {
		return db_common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		monthID, err := func() (uint, error) {
			in := &Income{ID: id}
			err = tx.Model(in).Column("month_id").WherePK().Select()
			if err != nil {
				return 0, err
			}

			return in.MonthID, nil
		}()
		if err != nil {
			return errors.Wrap(err, "couldn't select Month id of Income to remove")
		}

		// Remove income
		if err = db.removeIncome(tx, id); err != nil {
			return err
		}

		// Recompute Month budget
		if err := db.recomputeMonth(tx, monthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
}

func (DB) removeIncome(tx *pg.Tx, id uint) error {
	return tx.Delete(&Income{ID: id})
}
