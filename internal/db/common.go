package db

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/db/money"
)

// Errors

// Checks
var (
	ErrMonthNotExist          = badRequestError(errors.New("month with passed id doesn't exist"))
	ErrDayNotExist            = badRequestError(errors.New("day with passed id doesn't exist"))
	ErrIncomeNotExist         = badRequestError(errors.New("income with passed id doesn't exist"))
	ErrMonthlyPaymentNotExist = badRequestError(errors.New("monthly payment with passed id doesn't exist"))
	ErrSpendNotExist          = badRequestError(errors.New("spend with passed id doesn't exist"))
	ErrSpendTypeNotExist      = badRequestError(errors.New("spend type with passed id doesn't exist"))
)

// Common errors

func errBeginTransaction(err error) error {
	const msg = "can't begin a new transaction"
	return internalErrorWrap(err, msg)
}

func errCommitChanges(err error) error {
	const msg = "can't commit changes"
	return internalErrorWrap(err, msg)
}

func errRecomputeBudget(err error) error {
	const msg = "can't recompute month budget"
	return internalErrorWrap(err, msg)
}

// -----------------------------------------------------------------------------

type Month struct {
	ID    uint
	Year  int
	Month time.Month

	// Incomes

	Incomes     []*Income `pg:"fk:month_id"`
	TotalIncome money.Money
	// DailyBudget is a (TotalIncome - Cost of Monthly Payments) / Number of Days
	DailyBudget money.Money

	// Spends

	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`
	Days            []*Day            `pg:"fk:month_id"`
	TotalSpend      money.Money       // must be negative or zero

	// Result is TotalIncome - TotalSpend
	Result money.Money
}

func (db DB) GetMonth(id uint) (*Month, error) {
	m := &Month{ID: id}
	err := db.db.Model(m).
		Relation("Incomes").
		Relation("MonthlyPayments").
		Relation("Days", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("day ASC"), nil
		}).
		Relation("Days.Spends").
		WherePK().Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrMonthNotExist
		}
		return nil, internalErrorWrap(err, "can't select month with passed id")
	}

	return m, nil
}

type Day struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint

	Day int
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  money.Money
	Spends []*Spend `pg:"fk:day_id"`
}

func (db DB) GetDay(id uint) (*Day, error) {
	d := &Day{ID: id}
	err := db.db.Model(d).
		Relation("Spends").
		WherePK().
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrDayNotExist
		}
		return nil, internalErrorWrap(err, "can't select day with passed id")
	}

	return d, nil
}

// -----------------------------------------------------------------------------

func (db DB) GetMonthID(year, month int) (uint, error) {
	m := &Month{}
	err := db.db.Model(m).Column("id").Where("year = ? AND month = ?", year, month).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, ErrMonthNotExist
		}

		return 0, internalErrorWrap(err, "can't select Month with passed year and month")
	}

	return m.ID, nil
}

func (db DB) GetMonthIDByDay(dayID uint) (uint, error) {
	day := &Day{ID: dayID}
	err := db.db.Model(day).Column("month_id").WherePK().Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, ErrDayNotExist
		}

		return 0, internalErrorWrap(err, "can't select day with passed id")
	}

	return day.MonthID, nil
}

// Internal methods

func (_ DB) recomputeMonth(tx *pg.Tx, monthID uint) error {
	m := &Month{ID: monthID}
	err := tx.Model(m).
		Relation("Incomes").
		Relation("MonthlyPayments").
		Relation("Days").
		Relation("Days.Spends").
		WherePK().
		Select()
	if err != nil {
		return internalErrorWrap(err, "can't select month")
	}

	// Update Total Income
	m.TotalIncome = func() money.Money {
		var income money.Money
		for _, in := range m.Incomes {
			income = income.Add(in.Income)
		}
		return income
	}()

	// Update Total Spends and Daily Budget

	monthlyPaymentCost := func() money.Money {
		var cost money.Money
		for _, mp := range m.MonthlyPayments {
			cost = cost.Sub(mp.Cost)
		}
		return cost
	}()
	spendCost := func() money.Money {
		var cost money.Money
		for _, day := range m.Days {
			if day == nil {
				continue
			}
			for _, spend := range day.Spends {
				if spend == nil {
					continue
				}
				cost = cost.Sub(spend.Cost)
			}
		}
		return cost
	}()

	m.TotalSpend = monthlyPaymentCost.Add(spendCost)
	// Use "Add" because TotalSpend is negative
	m.DailyBudget = m.TotalIncome.Add(m.TotalSpend).Divide(int64(daysInMonth(m.Month)))

	// Update Saldos (it is accumulated)
	saldo := m.DailyBudget
	for i := range m.Days {
		if m.Days[i] == nil {
			continue
		}

		m.Days[i].Saldo = saldo
		for _, spend := range m.Days[i].Spends {
			if spend == nil {
				continue
			}
			m.Days[i].Saldo = m.Days[i].Saldo.Sub(spend.Cost)
		}
		saldo = m.Days[i].Saldo + m.DailyBudget
	}

	// Update Month
	err = tx.Update(m)
	if err != nil {
		return errorWrap(err, "can't update month")
	}

	// Update Days
	_, err = tx.Model(&m.Days).Update()
	if err != nil {
		return errorWrap(err, "can't update days")
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
