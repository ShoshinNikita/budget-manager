package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"

	. "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/models"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// -----------------------------------------------------------------------------
// Month
// -----------------------------------------------------------------------------

func (db DB) GetMonth(_ context.Context, id uint) (m *models.Month, err error) {
	err = db.db.RunInTransaction(func(tx *pg.Tx) error {
		m, err = db.getMonth(tx, id)
		return err
	})
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrMonthNotExist
		}
		return nil, errors.Wrap(err,
			errors.WithMsg("can't select month with passed id"),
			errors.WithType(errors.AppError))
	}

	return m, nil
}

func (db DB) GetMonthID(_ context.Context, year, month int) (uint, error) {
	m := &models.Month{}
	err := db.db.Model(m).Column("id").Where("year = ? AND month = ?", year, month).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, ErrMonthNotExist
		}

		return 0, errors.Wrap(err,
			errors.WithMsg("can't select Month with passed year and month"),
			errors.WithType(errors.AppError))
	}

	return m.ID, nil
}

func (DB) getMonthIDByDayID(_ context.Context, tx *pg.Tx, dayID uint) (uint, error) {
	day := &models.Day{ID: dayID}
	err := tx.Model(day).Column("month_id").WherePK().Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, ErrDayNotExist
		}

		return 0, errors.Wrap(err,
			errors.WithMsg("can't select day with passed id"),
			errors.WithType(errors.AppError))
	}

	return day.MonthID, nil
}

// GetMonths returns months of passed year. Months doesn't contains
// relations (Incomes, Days and etc.)
func (db DB) GetMonths(_ context.Context, year int) ([]*models.Month, error) {
	months := []*models.Month{}
	err := db.db.Model(&months).
		Where("year = ?", year).
		Order("month ASC").
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrYearNotExist
		}
		return nil, errors.Wrap(err,
			errors.WithMsg("can't select months with passed year"),
			errors.WithType(errors.AppError))
	}
	if len(months) == 0 {
		return nil, ErrYearNotExist
	}

	return months, nil
}

// -----------------------------------------------------------------------------
// Day
// -----------------------------------------------------------------------------

func (db DB) GetDay(_ context.Context, id uint) (*models.Day, error) {
	d := &models.Day{ID: id}
	err := db.db.Model(d).
		Relation("Spends", orderByID).
		Relation("Spends.Type").
		WherePK().Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrDayNotExist
		}
		return nil, errors.Wrap(err,
			errors.WithMsg("can't select day with passed id"),
			errors.WithType(errors.AppError))
	}

	return d, nil
}

func (db DB) GetDayIDByDate(ctx context.Context, year int, month int, day int) (uint, error) {
	monthID, err := db.GetMonthID(ctx, year, month)
	if err != nil {
		if err == ErrMonthNotExist {
			return 0, ErrMonthNotExist
		}
		return 0, errors.Wrap(err,
			errors.WithMsg("can't define month id with passed year and month"),
			errors.WithType(errors.AppError))
	}

	d := &models.Day{}
	err = db.db.Model(d).
		Column("id").
		Where("month_id = ? AND day = ?", monthID, day).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, ErrDayNotExist
		}
		return 0, errors.Wrap(err,
			errors.WithMsg("can't select day with passed id"),
			errors.WithType(errors.AppError))
	}

	return d.ID, nil
}

// -----------------------------------------------------------------------------
// Internal methods
// -----------------------------------------------------------------------------

// nolint:funlen
func (db DB) recomputeMonth(tx *pg.Tx, monthID uint) error {
	m, err := db.getMonth(tx, monthID)
	if err != nil {
		return errors.Wrap(err,
			errors.WithMsg("can't select month"),
			errors.WithType(errors.AppError))
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
	m.DailyBudget = m.TotalIncome.Add(monthlyPaymentCost).Divide(int64(daysInMonth(date)))
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
		return errors.Wrap(err,
			errors.WithMsg("can't update month"),
			errors.WithType(errors.AppError))
	}

	// Update Days
	_, err = tx.Model(&m.Days).Update()
	if err != nil {
		return errors.Wrap(err,
			errors.WithMsg("can't update days"),
			errors.WithType(errors.AppError))
	}

	return nil
}

func (DB) getMonth(tx *pg.Tx, id uint) (*models.Month, error) {
	m := &models.Month{ID: id}
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

	return m, err
}

func orderByID(q *orm.Query) (*orm.Query, error) {
	return q.Order("id ASC"), nil
}

// Checks

// checkMonth checks if Month with passed id exists
func (db DB) checkMonth(id uint) (ok bool) {
	m := &models.Month{ID: id}
	return db.checkModel(m)
}

// checkDay checks if Dat with passed id exists
func (db DB) checkDay(id uint) (ok bool) {
	d := &models.Day{ID: id}
	return db.checkModel(d)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkIncome(id uint) (ok bool) {
	st := &models.Income{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkMonthlyPayment(id uint) (ok bool) {
	st := &models.MonthlyPayment{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkSpend(id uint) (ok bool) {
	st := &models.Spend{ID: id}
	return db.checkModel(st)
}

// checkSpendType checks if Spend Type with passed id exists
func (db DB) checkSpendType(id uint) (ok bool) {
	st := &models.SpendType{ID: id}
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

// Other

func errRecomputeBudget(err error) error {
	return errors.Wrap(err,
		errors.WithMsg("can't recompute month budget"),
		errors.WithType(errors.AppError))
}
