package db

import (
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/pkg/money"
)

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	// MonthID is a foreign key to Months table
	MonthID uint `json:"month_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	Notes  string      `json:"notes,omitempty"`
	Income money.Money `json:"income"`
}

// Check checks whether Income is valid (not empty title, positive income and etc.)
func (in Income) Check() error {
	// Check Title
	if in.Title == "" {
		return badRequestError(errors.New("title can't be empty"))
	}

	// Skip Notes

	// Check Income
	if in.Income <= 0 {
		return badRequestError(errors.Errorf("invalid income: '%d'", in.Income))
	}

	return nil
}

// -----------------------------------------------------------------------------

type AddIncomeArgs struct {
	MonthID uint
	Title   string
	Notes   string
	Income  money.Money
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(args AddIncomeArgs) (incomeID uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, ErrMonthNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Add a new Income
		incomeID, err = db.addIncome(tx, args)
		if err != nil {
			err = errorWrap(err, "can't add income")
			db.log.Error(err)
			return err
		}

		// Recompute Month budget
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

	return incomeID, nil
}

func (DB) addIncome(tx *pg.Tx, args AddIncomeArgs) (incomeID uint, err error) {
	in := &Income{
		MonthID: args.MonthID,

		Title:  args.Title,
		Notes:  args.Notes,
		Income: args.Income,
	}

	if err := in.Check(); err != nil {
		return 0, err
	}
	err = tx.Insert(in)
	if err != nil {
		err = errorWrap(err, "can't insert a new row")
		return 0, err
	}

	return in.ID, nil
}

type EditIncomeArgs struct {
	ID     uint
	Title  *string
	Notes  *string
	Income *money.Money
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(args EditIncomeArgs) error {
	if !db.checkIncome(args.ID) {
		return ErrIncomeNotExist
	}

	in := &Income{ID: args.ID}

	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		// Select Income
		err = tx.Select(in)
		if err != nil {
			if err == pg.ErrNoRows {
				err = ErrIncomeNotExist
			} else {
				err = errorWrapf(err, "can't select Income with passed id '%d'", args.ID)
			}
			db.log.Error(err)
			return err
		}

		// Edit Income
		err = db.editIncome(tx, in, args)
		if err != nil {
			err = errorWrap(err, "can't edit Income")
			db.log.Error(err)
			return err
		}

		// Recompute Month budget
		err = db.recomputeMonth(tx, in.MonthID)
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

	if err := in.Check(); err != nil {
		return err
	}
	err := tx.Update(in)
	if err != nil {
		return errorWrap(err, "can't update Income")
	}

	return nil
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(id uint) error {
	if !db.checkIncome(id) {
		return ErrIncomeNotExist
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
			err = errorWrap(err, "can't select Income with passed id")
			db.log.Error(err)
			return err
		}

		// Remove income
		err = db.removeIncome(tx, id)
		if err != nil {
			err = errorWrap(err, "can't remove Income with passed id")
			db.log.Error(err)
			return err
		}

		// Recompute Month budget
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

func (DB) removeIncome(tx *pg.Tx, id uint) error {
	in := &Income{ID: id}
	return tx.Delete(in)
}
