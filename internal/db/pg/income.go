package pg

import (
	"context"

	"github.com/go-pg/pg/v9"

	. "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/models"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// AddIncome adds a new income with passed params
func (db DB) AddIncome(_ context.Context, args AddIncomeArgs) (incomeID uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, ErrMonthNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add a new Income
		incomeID, err = db.addIncome(tx, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't add a new Income"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute Month budget
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

	return incomeID, nil
}

func (DB) addIncome(tx *pg.Tx, args AddIncomeArgs) (incomeID uint, err error) {
	in := &models.Income{
		MonthID: args.MonthID,

		Title:  args.Title,
		Notes:  args.Notes,
		Income: args.Income,
	}

	if err := checkModel(in); err != nil {
		return 0, err
	}
	err = tx.Insert(in)
	if err != nil {
		return 0, err
	}

	return in.ID, nil
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(_ context.Context, args EditIncomeArgs) error {
	if !db.checkIncome(args.ID) {
		return ErrIncomeNotExist
	}

	in := &models.Income{ID: args.ID}

	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Select Income
		err = tx.Select(in)
		if err != nil {
			if err == pg.ErrNoRows {
				return ErrIncomeNotExist
			}
			return errors.Wrap(err, errors.WithMsg("can't select Income"), errors.WithType(errors.AppError))
		}

		// Edit Income
		err = db.editIncome(tx, in, args)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't edit the Income"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute Month budget
		err = db.recomputeMonth(tx, in.MonthID)
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

func (DB) editIncome(tx *pg.Tx, in *models.Income, args EditIncomeArgs) error {
	if args.Title != nil {
		in.Title = *args.Title
	}
	if args.Notes != nil {
		in.Notes = *args.Notes
	}
	if args.Income != nil {
		in.Income = *args.Income
	}

	if err := checkModel(in); err != nil {
		return err
	}
	err := tx.Update(in)
	if err != nil {
		return err
	}

	return nil
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(_ context.Context, id uint) error {
	if !db.checkIncome(id) {
		return ErrIncomeNotExist
	}

	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		monthID, err := func() (uint, error) {
			in := &models.Income{ID: id}
			err = tx.Model(in).Column("month_id").WherePK().Select()
			if err != nil {
				return 0, err
			}

			return in.MonthID, nil
		}()
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't select Income with passed id"),
				errors.WithType(errors.AppError))
		}

		// Remove income
		err = db.removeIncome(tx, id)
		if err != nil {
			return errors.Wrap(err,
				errors.WithMsg("can't remove Income with passed id"),
				errors.WithTypeIfNotSet(errors.AppError))
		}

		// Recompute Month budget
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

func (DB) removeIncome(tx *pg.Tx, id uint) error {
	in := &models.Income{ID: id}
	return tx.Delete(in)
}
