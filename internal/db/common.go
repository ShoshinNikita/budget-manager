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
	ID    uint
	Year  int
	Month time.Month

	// Incomes

	Incomes     []*Income `pg:"fk:month_id"`
	TotalIncome int64
	// DailyBudget is a (TotalIncome - Cost of Monthly Payments) / Number of Days
	// multiplied by 100
	DailyBudget int64

	// Spends

	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`
	Days            []*Day            `pg:"fk:month_id"`
	TotalSpend      int64             // must be negative or zero

	// Result is TotalIncome - TotalSpend
	Result int64
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

func (db DB) GetMonthIDByDay(dayID uint) (uint, error) {
	day := &Day{ID: dayID}
	err := db.db.Model(day).Column("month_id").WherePK().Select()
	if err != nil {
		return 0, errors.Wrap(err, "can't select day with passed id")
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
		return errors.Wrap(err, "can't select month")
	}

	// Update Total Income
	m.TotalIncome = func() int64 {
		var income int64
		for _, in := range m.Incomes {
			income += in.Income
		}
		return income
	}()

	// Update Total Spends and Daily Budget

	monthlyPaymentCost := func() int64 {
		var cost int64
		for _, mp := range m.MonthlyPayments {
			cost -= mp.Cost
		}
		return cost
	}()
	spendCost := func() int64 {
		var cost int64
		for _, day := range m.Days {
			if day == nil {
				continue
			}
			for _, spend := range day.Spends {
				if spend == nil {
					continue
				}
				cost -= spend.Cost
			}
		}
		return cost
	}()

	m.TotalSpend = monthlyPaymentCost + spendCost
	// Use '+' because monthlyPaymentsCost is negative
	m.DailyBudget = (m.TotalIncome + m.TotalSpend) / int64(daysInMonth(m.Month))

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
			m.Days[i].Saldo -= spend.Cost
		}
		saldo = m.Days[i].Saldo + m.DailyBudget
	}

	// Update Month
	err = tx.Update(m)
	if err != nil {
		return errors.Wrap(err, "can't update month")
	}

	// Update Days
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
