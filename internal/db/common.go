package db

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

// Errors

// Checks
var (
	ErrMonthNotExist          = errors.New("month with passed id doesn't exist")
	ErrDayNotExist            = errors.New("day with passed id doesn't exist")
	ErrIncomeNotExist         = errors.New("income with passed id doesn't exist")
	ErrMonthlyPaymentNotExist = errors.New("monthly payment with passed id doesn't exist")
	ErrSpendNotExist          = errors.New("spend with passed id doesn't exist")
	ErrSpendTypeNotExist      = errors.New("spend type with passed id doesn't exist")
)

// Transaction (wrap messages)
const (
	errBeginTransaction = "can't begin a new transaction"
	errCommitChanges    = "can't commit changes"
	errRecomputeBudget  = "can't recompute month budget"
)

// -----------------------------------------------------------------------------

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
		WherePK().
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

// Checks

// checkMonth checks if Month with passed id exists
func (db DB) checkMonth(id uint) (ok bool) {
	m := &Month{ID: id}
	return db.checkModel(m)
}

// checkDay checks if Dat with passed id exists
func (db DB) checkDay(id uint) (ok bool) {
	d := &Day{ID: id}
	return db.checkModel(d)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkIncome(id uint) (ok bool) {
	st := &Income{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkMonthlyPayment(id uint) (ok bool) {
	st := &MonthlyPayment{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkSpend(id uint) (ok bool) {
	st := &Spend{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkSpendType(id uint) (ok bool) {
	st := &SpendType{ID: id}
	return db.checkModel(st)
}

// checkModel checks if model with primary key exists
func (db DB) checkModel(model interface{}) (ok bool) {
	c, err := db.db.Model(model).WherePK().Count()
	if c == 0 || err != nil {
		return false
	}
	return true
}
