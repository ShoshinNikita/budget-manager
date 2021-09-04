package pg

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type MonthOverview struct {
	ID          uint        `db:"id"`
	Year        int         `db:"year"`
	Month       time.Month  `db:"month"`
	DailyBudget money.Money `db:"daily_budget"`
	TotalIncome money.Money `db:"total_income"`
	TotalSpend  money.Money `db:"total_spend"`
	Result      money.Money `db:"result"`
}

func (m MonthOverview) ToCommon() common.MonthOverview {
	return common.MonthOverview{
		ID:          m.ID,
		Year:        m.Year,
		Month:       m.Month,
		TotalIncome: m.TotalIncome,
		TotalSpend:  m.TotalSpend,
		DailyBudget: m.DailyBudget,
		Result:      m.Result,
	}
}

type Month struct {
	MonthOverview

	Incomes         []Income         `db:"-"`
	MonthlyPayments []MonthlyPayment `db:"-"`
	Days            []Day            `db:"-"`
}

func (m Month) ToCommon() common.Month {
	return common.Month{
		MonthOverview: m.MonthOverview.ToCommon(),
		//
		Incomes: func() []common.Income {
			incomes := make([]common.Income, 0, len(m.Incomes))
			for i := range m.Incomes {
				incomes = append(incomes, m.Incomes[i].ToCommon(m.Year, m.Month))
			}
			return incomes
		}(),
		MonthlyPayments: func() []common.MonthlyPayment {
			mp := make([]common.MonthlyPayment, 0, len(m.MonthlyPayments))
			for i := range m.MonthlyPayments {
				mp = append(mp, m.MonthlyPayments[i].ToCommon(m.Year, m.Month))
			}
			return mp
		}(),
		Days: func() []common.Day {
			days := make([]common.Day, 0, len(m.Days))
			for i := range m.Days {
				days = append(days, m.Days[i].ToCommon(m.Year, m.Month))
			}
			return days
		}(),
	}
}

func (db DB) GetMonthByDate(ctx context.Context, year int, month time.Month) (common.Month, error) {
	var m Month
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) (err error) {
		m, err = getFullMonth(tx, "year = ? AND month = ?", year, month)
		return err
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = common.ErrMonthNotExist
		}
		return common.Month{}, err
	}

	return m.ToCommon(), nil
}

// GetMonths returns months of passed years. Months doesn't contains relations (Incomes, Days and etc.)
func (db DB) GetMonths(ctx context.Context, years ...int) ([]common.MonthOverview, error) {
	var m []MonthOverview
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		return tx.SelectQuery(&m, sqlx.In(`SELECT * FROM months WHERE year IN (?) ORDER BY id ASC`, years))
	})
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}

	res := make([]common.MonthOverview, 0, len(m))
	for i := range m {
		res = append(res, m[i].ToCommon())
	}
	return res, nil
}

// InitMonth inits a month and days for the passed date
func (db *DB) InitMonth(ctx context.Context, year int, month time.Month) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		var count int
		err := tx.Get(&count, `SELECT COUNT(*) FROM months WHERE year = ? AND month = ?`, year, month)
		if err != nil {
			return errors.Wrap(err, "couldn't check if the current month exists")
		}
		if count != 0 {
			// The month is already created
			return nil
		}

		// We have to init the current month

		var monthID uint
		err = tx.Get(&monthID, `INSERT INTO months(year, month) VALUES(?, ?) RETURNING id`, year, month)
		if err != nil {
			return errors.Wrap(err, "couldn't init the current month")
		}

		query := squirrel.Insert("days").Columns("month_id", "day")

		daysNumber := daysInMonth(year, month)
		for i := 0; i < daysNumber; i++ {
			query = query.Values(monthID, i+1)
		}
		if _, err = tx.ExecQuery(query); err != nil {
			return errors.Wrap(err, "couldn't insert days for the current month")
		}
		return nil
	})
}

func (db DB) recomputeAndUpdateMonth(tx *sqlx.Tx, monthID uint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "couldn't recompute the month budget")
		}
	}()

	m, err := getFullMonth(tx, "id = ?", monthID)
	if err != nil {
		return errors.Wrap(err, "couldn't select month")
	}

	m = recomputeMonth(m)

	// Update Month
	_, err = tx.Exec(
		`UPDATE months SET daily_budget = ?, total_income = ?, total_spend = ?, result = ? WHERE id = ?`,
		m.DailyBudget, m.TotalIncome, m.TotalSpend, m.Result, m.ID,
	)

	// Update Days

	values := &bytes.Buffer{}
	for i, day := range m.Days {
		fmt.Fprintf(values, `(%d,%d)`, day.ID, int(day.Saldo))
		if i+1 != len(m.Days) {
			values.WriteByte(',')
		}
	}

	_, err = tx.Exec(
		fmt.Sprintf(
			`UPDATE days SET saldo = v.saldo FROM (values %s) AS v(id, saldo) WHERE days.id = v.id`,
			values.String(),
		),
	)
	if err != nil {
		return errors.Wrap(err, "couldn't update days")
	}

	return nil
}

func recomputeMonth(m Month) Month {
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
		for _, spend := range day.Spends {
			spendsCost = spendsCost.Sub(spend.Cost)
		}
	}

	// Use "Add" because monthlyPaymentCost and TotalSpend are negative
	m.DailyBudget = m.TotalIncome.Add(monthlyPaymentsCost).Div(int64(len(m.Days)))
	m.TotalSpend = monthlyPaymentsCost.Add(spendsCost)
	m.Result = m.TotalIncome.Add(m.TotalSpend)

	// Update Saldos (it is accumulated)
	saldo := m.DailyBudget
	for i := range m.Days {
		m.Days[i].Saldo = saldo
		for _, spend := range m.Days[i].Spends {
			m.Days[i].Saldo = m.Days[i].Saldo.Sub(spend.Cost)
		}
		saldo = m.Days[i].Saldo + m.DailyBudget
	}

	return m
}

func getFullMonth(tx *sqlx.Tx, whereCond string, args ...interface{}) (m Month, err error) {
	err = tx.Get(&m.MonthOverview, `SELECT * from months WHERE `+whereCond, args...)
	if err != nil {
		return Month{}, errors.Wrap(err, "couldn't select month")
	}

	err = tx.Select(&m.Incomes, `SELECT * FROM incomes WHERE month_id = ? ORDER BY id`, m.ID)
	if err != nil {
		return Month{}, errors.Wrap(err, "couldn't select incomes")
	}
	err = tx.Select(
		&m.MonthlyPayments, `
		SELECT
			monthly_payments.*,
			spend_types.id AS "type.id",
			spend_types.name AS "type.name",
			spend_types.parent_id AS "type.parent_id"
		FROM monthly_payments
		LEFT JOIN spend_types ON spend_types.id = monthly_payments.type_id
		WHERE monthly_payments.month_id = ?
		ORDER BY monthly_payments.id`, m.ID,
	)
	if err != nil {
		return Month{}, errors.Wrap(err, "couldn't select monthly payments")
	}

	err = tx.Select(&m.Days, `SELECT * FROM days WHERE month_id = ? ORDER BY day`, m.ID)
	if err != nil {
		return Month{}, errors.Wrap(err, "couldn't select days")
	}

	dayIndexes := make(map[uint]int) // day id -> slice index
	dayIDs := make([]int, 0, len(m.Days))
	for i, d := range m.Days {
		dayIndexes[d.ID] = i
		dayIDs = append(dayIDs, int(d.ID))
	}
	var allSpends []Spend
	err = tx.SelectQuery(&allSpends, sqlx.In(`
		SELECT
			spends.*,
			spend_types.id AS "type.id",
			spend_types.name AS "type.name",
			spend_types.parent_id AS "type.parent_id"
		FROM spends
		LEFT JOIN spend_types ON spend_types.id = spends.type_id
		WHERE spends.day_id IN (?)
		ORDER BY spends.id`, dayIDs,
	))
	if err != nil {
		return Month{}, errors.Wrap(err, "couldn't select spends")
	}
	for _, s := range allSpends {
		dayIndex, ok := dayIndexes[s.DayID]
		if !ok {
			// Just in case
			return Month{}, errors.Errorf("spend with id %d has unexpected day id: %d", s.ID, s.DayID)
		}
		m.Days[dayIndex].Spends = append(m.Days[dayIndex].Spends, s)
	}

	return m, nil
}
