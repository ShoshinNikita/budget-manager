package db

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

var (
	errBeginTransaction = "can't begin a new transaction"
	errCommitChanges    = "can't commit changes"
	errRecomputeBudget  = "can't recompute month budget"
)

type Month struct {
	ID uint

	Year  int
	Month time.Month

	Incomes         []*Income         `pg:"fk:month_id"`
	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`

	// DailyBudget is a (Sum of Incomes - Cost of Monthly Payments) / Number of Days
	// multiplied by 100
	DailyBudget int64
	Days        []*Day `pg:"fk:month_id"`
}

type Day struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint

	Day int
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  int64
	Spends []*Spend `pg:"fk:day_id"`
}

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

// Internal methods

func (_ DB) recomputeMonth(tx *pg.Tx, monthID uint) error {
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
