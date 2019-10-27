package db

import (
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

var (
	errBeginTransaction = "can't begin a new transaction"
	errCommitChanges    = "can't commit changes"
)

// -----------------------------------------------------------------------------
// Income
// -----------------------------------------------------------------------------

type AddIncomeArgs struct {
	MonthID uint
	Title   string
	Notes   string
	Income  int64
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(args AddIncomeArgs) (incomeID uint, err error) {
	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		db.log.Error(err)
		return 0, err
	}
	defer tx.Rollback()

	// Add a new Income
	id, err := db.addIncome(tx, args)
	if err != nil {
		err = errors.Wrap(err, "can't add income")
		db.log.Error(err)
		return 0, err
	}

	// Recompute Month budget
	err = recomputeMonth(tx, args.MonthID)
	if err != nil {
		err = errors.Wrap(err, "can't recompute month budget")
		db.log.Error(err)
		return 0, err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return 0, err
	}

	return id, nil
}

func (_ DB) addIncome(tx *pg.Tx, args AddIncomeArgs) (incomeID uint, err error) {
	in := &Income{
		MonthID: args.MonthID,

		Title:  args.Title,
		Notes:  args.Notes,
		Income: args.Income,
	}

	err = tx.Insert(in)
	if err != nil {
		err = errors.Wrap(err, "can't insert a new row")
		return 0, err
	}

	return in.ID, nil
}

type EditIncomeArgs struct {
	ID     uint
	Title  *string
	Notes  *string
	Income *int64
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(args EditIncomeArgs) error {
	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		db.log.Error(err)
		return err
	}
	defer tx.Rollback()

	// Select Income.
	in := &Income{ID: args.ID}
	err = tx.Select(in)
	if err != nil {
		if err == pg.ErrNoRows {
			err = errors.New("wrong id")
		} else {
			err = errors.Wrapf(err, "can't select Income with passed id '%d'", args.ID)
		}
		db.log.Error(err)
		return err
	}

	// Edit Income
	err = db.editIncome(tx, in, args)
	if err != nil {
		err = errors.Wrap(err, "can't edit Income")
		db.log.Error(err)
		return err
	}

	// Recompute Month budget
	err = recomputeMonth(tx, in.MonthID)
	if err != nil {
		err = errors.Wrap(err, "can't recompute month budget")
		db.log.Error(err)
		return err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return err
	}

	return nil
}

func (_ DB) editIncome(tx *pg.Tx, in *Income, args EditIncomeArgs) error {
	if args.Title != nil {
		in.Title = *args.Title
	}
	if args.Notes != nil {
		in.Notes = *args.Notes
	}
	if args.Income != nil {
		in.Income = *args.Income
	}

	err := tx.Update(in)
	if err != nil {
		return errors.Wrap(err, "can't update Income")
	}

	return nil
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(id uint) error {
	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		return err
	}
	defer tx.Rollback()

	monthID, err := func() (uint, error) {
		in := &Income{ID: id}
		err = tx.Model(in).Column("month_id").WherePK().Select(in)
		if err != nil {
			return 0, err
		}

		return in.MonthID, nil
	}()
	if err != nil {
		err = errors.Wrap(err, "can't select Income with passed id")
		db.log.Error(err)
		return err
	}

	// Remove income
	err = db.removeIncome(tx, id)
	if err != nil {
		err = errors.Wrap(err, "can't remove Income with passed id")
		db.log.Error(err)
		return err
	}

	// Recompute Month budget
	err = recomputeMonth(tx, monthID)
	if err != nil {
		err = errors.Wrap(err, "can't recompute month budget")
		db.log.Error(err)
		return err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return err
	}

	return nil
}

func (_ DB) removeIncome(tx *pg.Tx, id uint) error {
	in := &Income{ID: id}
	return tx.Delete(in)
}

// -----------------------------------------------------------------------------
// Monthly Payments
// -----------------------------------------------------------------------------

func (db DB) AddMonthlyPayment() {}

func (db DB) EditMonthlyPayment() {}

func (db DB) RemoveMonthlyPayment() {}

// -----------------------------------------------------------------------------
// Spends
// -----------------------------------------------------------------------------

func (db DB) AddSpend() {}

func (db DB) EditSpend() {}

func (db DB) RemoveSpend() {}

// -----------------------------------------------------------------------------
// Spend Types
// -----------------------------------------------------------------------------

func (db DB) AddSpendType() {}

func (db DB) EditSpendType() {}

func (db DB) RemoveSpendType() {}

// -----------------------------------------------------------------------------

func (db DB) GetMonthID(year, month int) (uint, error) {
	m := &Month{}
	err := db.db.Model(m).Column("id").Where("year = ? AND month = ?", year, month).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, errors.New("there is no such month")
		}

		return 0, errors.Wrap(err, "can't select Month with passed year and month")
	}

	return m.ID, nil
}

// -----------------------------------------------------------------------------

func recomputeMonth(tx *pg.Tx, monthID uint) error {
	m := &Month{ID: monthID}
	err := tx.Model(m).
		Relation("Incomes").
		Relation("MonthlyPayments").
		Relation("Days").
		Select()

	if err != nil {
		return errors.Wrap(err, "can't select month")
	}

	newDailyBudget := func() int64 {
		var monthlyBudget int64

		for _, in := range m.Incomes {
			monthlyBudget += in.Income
		}
		for _, p := range m.MonthlyPayments {
			monthlyBudget -= p.Cost
		}

		return monthlyBudget / int64(daysInMonth(m.Month))
	}()

	// deltaDailyBudget is used to update saldo. deltaDailyBudget can be negative
	deltaDailyBudget := newDailyBudget - m.DailyBudget

	// Update daily budget
	m.DailyBudget = newDailyBudget

	// Update Saldo
	for _, day := range m.Days {
		day.Saldo += deltaDailyBudget
	}

	// Update Month
	err = tx.Update(m)
	if err != nil {
		return errors.Wrap(err, "can't update month")
	}

	_, err = tx.Model(&m.Days).Update()
	if err != nil {
		return errors.Wrap(err, "can't update days")
	}

	return nil
}
