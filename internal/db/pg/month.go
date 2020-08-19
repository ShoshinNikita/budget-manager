package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func (db DB) GetMonth(_ context.Context, id uint) (*db_common.Month, error) {
	var pgMonth *Month
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		pgMonth, err = db.getMonth(tx, id)
		return err
	})
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, db_common.ErrMonthNotExist
		}
		return nil, err
	}

	return pgMonth.ToCommon(), nil
}

func (db DB) GetMonthID(_ context.Context, year, month int) (id uint, err error) {
	err = db.db.Model((*Month)(nil)).Column("id").Where("year = ? AND month = ?", year, month).Select(&id)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, db_common.ErrMonthNotExist
		}
		return 0, err
	}

	return id, nil
}

// GetMonths returns months of passed year. Months doesn't contains
// relations (Incomes, Days and etc.)
func (db DB) GetMonths(_ context.Context, year int) ([]*db_common.Month, error) {
	var pgMonths []*Month
	err := db.db.Model(&pgMonths).Where("year = ?", year).Order("month ASC").Select()
	if err != nil {
		return nil, err
	}
	if len(pgMonths) == 0 {
		return nil, db_common.ErrYearNotExist
	}

	months := make([]*db_common.Month, 0, len(pgMonths))
	for i := range pgMonths {
		months = append(months, pgMonths[i].ToCommon())
	}
	return months, nil
}

// nolint:funlen
func (db DB) recomputeMonth(tx *pg.Tx, monthID uint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "couldn't recompute the month budget")
		}
	}()

	m, err := db.getMonth(tx, monthID)
	if err != nil {
		return errors.Wrap(err, "couldn't select month")
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

	date := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.Local)
	// Use "Add" because monthlyPaymentCost and TotalSpend are negative
	m.DailyBudget = m.TotalIncome.Add(monthlyPaymentCost).Divide(int64(daysInMonth(date.Year(), date.Month())))
	m.TotalSpend = monthlyPaymentCost.Add(spendCost)
	m.Result = m.TotalIncome.Add(m.TotalSpend)

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
		return errors.Wrap(err, "couldn't update month")
	}

	// Update Days
	_, err = tx.Model(&m.Days).Update()
	if err != nil {
		return errors.Wrap(err, "couldn't update days")
	}

	return nil
}

func (DB) getMonth(tx *pg.Tx, id uint) (*Month, error) {
	m := &Month{ID: id}
	err := tx.Model(m).
		Relation("Incomes", orderByID).
		Relation("MonthlyPayments", orderByID).
		Relation("MonthlyPayments.Type").
		Relation("Days", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("day ASC"), nil
		}).
		Relation("Days.Spends", orderByID).
		Relation("Days.Spends.Type").
		WherePK().Select()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func orderByID(q *orm.Query) (*orm.Query, error) {
	return q.Order("id ASC"), nil
}
