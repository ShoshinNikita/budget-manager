package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// Month represents month entity in PostgreSQL db
type Month struct {
	tableName struct{} `pg:"months"` // nolint:structcheck,unused

	ID uint `pg:"id,pk"`

	Year  int        `pg:"year"`
	Month time.Month `pg:"month"`

	Incomes         []*Income         `pg:"fk:month_id"`
	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`

	// DailyBudget is a (TotalIncome - Cost of Monthly Payments) / Number of Days
	DailyBudget money.Money `pg:"daily_budget,use_zero"`
	Days        []*Day      `pg:"fk:month_id"`

	TotalIncome money.Money `pg:"total_income,use_zero"`
	// TotalSpend is a cost of all Monthly Payments and Spends
	TotalSpend money.Money `pg:"total_spend,use_zero"`
	// Result is TotalIncome - TotalSpend
	Result money.Money `pg:"result,use_zero"`
}

// ToCommon converts Month to common Month structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (m *Month) ToCommon() *common.Month {
	if m == nil {
		return nil
	}
	return &common.Month{
		ID:          m.ID,
		Year:        m.Year,
		Month:       m.Month,
		TotalIncome: m.TotalIncome,
		TotalSpend:  m.TotalSpend,
		DailyBudget: m.DailyBudget,
		Result:      m.Result,
		//
		Incomes: func() []*common.Income {
			incomes := make([]*common.Income, 0, len(m.Incomes))
			for i := range m.Incomes {
				incomes = append(incomes, m.Incomes[i].ToCommon(m.Year, m.Month))
			}
			return incomes
		}(),
		MonthlyPayments: func() []*common.MonthlyPayment {
			mp := make([]*common.MonthlyPayment, 0, len(m.MonthlyPayments))
			for i := range m.MonthlyPayments {
				mp = append(mp, m.MonthlyPayments[i].ToCommon(m.Year, m.Month))
			}
			return mp
		}(),
		Days: func() []*common.Day {
			days := make([]*common.Day, 0, len(m.Days))
			for i := range m.Days {
				days = append(days, m.Days[i].ToCommon(m.Year, m.Month))
			}
			return days
		}(),
	}
}

func (db DB) GetMonth(_ context.Context, id uint) (*common.Month, error) {
	var pgMonth *Month
	err := db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		pgMonth, err = db.getMonth(tx, id)
		return err
	})
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, common.ErrMonthNotExist
		}
		return nil, err
	}

	return pgMonth.ToCommon(), nil
}

func (db DB) GetMonthID(_ context.Context, year, month int) (id uint, err error) {
	err = db.db.Model((*Month)(nil)).Column("id").Where("year = ? AND month = ?", year, month).Select(&id)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, common.ErrMonthNotExist
		}
		return 0, err
	}

	return id, nil
}

// GetMonths returns months of passed year. Months doesn't contains
// relations (Incomes, Days and etc.)
func (db DB) GetMonths(_ context.Context, year int) ([]*common.Month, error) {
	var pgMonths []*Month
	err := db.db.Model(&pgMonths).Where("year = ?", year).Order("month ASC").Select()
	if err != nil {
		return nil, err
	}
	if len(pgMonths) == 0 {
		return nil, common.ErrYearNotExist
	}

	months := make([]*common.Month, 0, len(pgMonths))
	for i := range pgMonths {
		months = append(months, pgMonths[i].ToCommon())
	}
	return months, nil
}

// Internal methods

// initCurrentMonth inits month for current year and month
func (db *DB) initCurrentMonth() error {
	year, month, _ := time.Now().Date()
	return db.initMonth(year, month)
}

// initMonth inits month and days for passed date
func (db *DB) initMonth(year int, month time.Month) error {
	count, err := db.db.Model((*Month)(nil)).Where("year = ? AND month = ?", year, month).Count()
	if err != nil {
		return errors.Wrap(err, "couldn't check if the current month exists")
	}
	if count != 0 {
		// The month is already created
		return nil
	}

	// We have to init the current month

	log := db.log

	// Add the current month
	log.Debug("init the current month")

	currentMonth := &Month{Year: year, Month: month}
	if _, err = db.db.Model(currentMonth).Returning("id").Insert(); err != nil {
		return errors.Wrap(err, "couldn't init the current month")
	}

	monthID := currentMonth.ID
	log = log.WithField("month_id", monthID)
	log.Debug("current month was successfully inited")

	// Add days for the current month
	log.Debug("init days of the current month")

	daysNumber := daysInMonth(year, month)
	days := make([]*Day, daysNumber)
	for i := range days {
		days[i] = &Day{MonthID: monthID, Day: i + 1, Saldo: 0}
	}

	if err = db.db.Insert(&days); err != nil {
		return errors.Wrap(err, "couldn't insert days for the current month")
	}
	log.Debug("days of the current month was successfully inited")

	return nil
}

func (db DB) recomputeAndUpdateMonth(tx *pg.Tx, monthID uint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "couldn't recompute the month budget")
		}
	}()

	m, err := db.getMonth(tx, monthID)
	if err != nil {
		return errors.Wrap(err, "couldn't select month")
	}

	m = recomputeMonth(m)

	// Update Month
	query := tx.Model((*Month)(nil)).Where("id = ?", m.ID)
	query = query.Set("daily_budget = ?", m.DailyBudget)
	query = query.Set("total_income = ?", m.TotalIncome)
	query = query.Set("total_spend = ?", m.TotalSpend)
	query = query.Set("result = ?", m.Result)
	if _, err := query.Update(); err != nil {
		return errors.Wrap(err, "couldn't update month")
	}

	// Update Days
	for _, day := range m.Days {
		query := tx.Model((*Day)(nil)).Where("id = ?", day.ID).Set("saldo = ?", day.Saldo)
		if _, err := query.Update(); err != nil {
			return errors.Wrap(err, "couldn't update days")
		}
	}

	return nil
}

func recomputeMonth(m *Month) *Month {
	// Update Total Income
	m.TotalIncome = 0
	for _, in := range m.Incomes {
		m.TotalIncome = m.TotalIncome.Add(in.Income)
	}

	// Update Total Spends and Daily Budget

	var monthlyPaymentsCost money.Money
	for _, mp := range m.MonthlyPayments {
		monthlyPaymentsCost = monthlyPaymentsCost.Sub(mp.Cost)
	}

	var spendsCost money.Money
	for _, day := range m.Days {
		if day == nil {
			continue
		}
		for _, spend := range day.Spends {
			if spend == nil {
				continue
			}
			spendsCost = spendsCost.Sub(spend.Cost)
		}
	}

	date := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.Local)
	daysNumber := daysInMonth(date.Year(), date.Month())

	// Use "Add" because monthlyPaymentCost and TotalSpend are negative
	m.DailyBudget = m.TotalIncome.Add(monthlyPaymentsCost).Divide(int64(daysNumber))
	m.TotalSpend = monthlyPaymentsCost.Add(spendsCost)
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

	return m
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
