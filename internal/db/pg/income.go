package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// AddIncome adds a new income with passed params
func (db DB) AddIncome(ctx context.Context, args db_common.AddIncomeArgs) (incomeID uint, err error) {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{
		"month_id": args.MonthID, "title": args.Title, "income": args.Income,
	})

	if !db.checkMonth(args.MonthID) {
		err := db_common.ErrMonthNotExist
		log.Error(err)
		return 0, err
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
		log.WithError(err).Error("couldn't add a new Income")
		return 0, err
	}

	log.WithField("id", incomeID).Debug("a new Income was successfully created")
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
func (db DB) EditIncome(ctx context.Context, args db_common.EditIncomeArgs) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{
		"id":         args.ID,
		"new_title":  args.Title,
		"new_income": args.Income,
	})

	if !db.checkIncome(args.ID) {
		err := db_common.ErrIncomeNotExist
		log.Error(err)
		return err
	}

	in := &Income{ID: args.ID}
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Select Income
		err = tx.Select(in)
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return db_common.ErrIncomeNotExist
			}
			return errors.Wrap(err, "couldn't select Income")
		}

		// Edit Income
		if err = db.editIncome(tx, in, args); err != nil {
			return err
		}

		// Recompute Month budget
		if err = db.recomputeMonth(tx, in.MonthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't edit Income")
		return err
	}

	log.Debug("the Income was successfully edited")
	return nil
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
func (db DB) RemoveIncome(ctx context.Context, id uint) error {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithField("id", id)

	if !db.checkIncome(id) {
		err := db_common.ErrIncomeNotExist
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
			return errors.Wrap(err, "couldn't select month id of Income with passed id")
		}

		// Remove income
		if err = db.removeIncome(tx, id); err != nil {
			return err
		}

		// Recompute Month budget
		if err = db.recomputeMonth(tx, monthID); err != nil {
			return errRecomputeBudget(err)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("couldn't remove Income")
		return err
	}

	log.Debug("the Income was successfully removed")
	return nil
}

func (DB) removeIncome(tx *pg.Tx, id uint) error {
	return tx.Delete(&Income{ID: id})
}
