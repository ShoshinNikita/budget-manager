package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"

	. "github.com/ShoshinNikita/budget-manager/internal/db" // nolint:stylecheck,golint
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// AddIncome adds a new income with passed params
func (db DB) AddIncome(_ context.Context, args AddIncomeArgs) (incomeID uint, err error) {
	log := db.log.WithFields(logrus.Fields{
		"month_id": args.MonthID,
		"title":    args.Title,
		"income":   args.Income,
	})

	if !db.checkMonth(args.MonthID) {
		err := ErrMonthNotExist
		log.Error(err)
		return 0, err
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
		log.WithError(err).Error("couldn't add a new Income")
		return 0, err
	}

	log.WithField("id", incomeID).Info("a new Income was successfully created")
	return incomeID, nil
}

func (DB) addIncome(tx *pg.Tx, args AddIncomeArgs) (incomeID uint, err error) {
	in := &Income{
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
	log := db.log.WithFields(logrus.Fields{
		"id":         args.ID,
		"new_title":  args.Title,
		"new_income": args.Income,
	})

	if !db.checkIncome(args.ID) {
		err := ErrIncomeNotExist
		log.Error(err)
		return err
	}

	in := &Income{ID: args.ID}

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
		log.WithError(err).Error("couldn't edit Income")
		return err
	}

	log.Info("the Income was successfully edited")
	return nil
}

func (DB) editIncome(tx *pg.Tx, in *Income, args EditIncomeArgs) error {
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
	log := db.log.WithField("id", id)

	if !db.checkIncome(id) {
		err := ErrIncomeNotExist
		log.Error(err)
		return err
	}

	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		monthID, err := func() (uint, error) {
			in := &Income{ID: id}
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
		log.WithError(err).Error("coudln't remove Income")
		return err
	}

	log.Info("the Income was successfully removed")
	return nil
}

func (DB) removeIncome(tx *pg.Tx, id uint) error {
	in := &Income{ID: id}
	return tx.Delete(in)
}
