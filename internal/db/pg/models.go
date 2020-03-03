package pg

import (
	"errors"
	"fmt"
	"time"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
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
func (m *Month) ToCommon() *db_common.Month {
	if m == nil {
		return nil
	}
	return &db_common.Month{
		ID:          m.ID,
		Year:        m.Year,
		Month:       m.Month,
		TotalIncome: m.TotalIncome,
		TotalSpend:  m.TotalSpend,
		DailyBudget: m.DailyBudget,
		Result:      m.Result,
		//
		Incomes: func() []*db_common.Income {
			incomes := make([]*db_common.Income, 0, len(m.Incomes))
			for i := range m.Incomes {
				incomes = append(incomes, m.Incomes[i].ToCommon(m.Year, m.Month))
			}
			return incomes
		}(),
		MonthlyPayments: func() []*db_common.MonthlyPayment {
			mp := make([]*db_common.MonthlyPayment, 0, len(m.MonthlyPayments))
			for i := range m.MonthlyPayments {
				mp = append(mp, m.MonthlyPayments[i].ToCommon(m.Year, m.Month))
			}
			return mp
		}(),
		Days: func() []*db_common.Day {
			days := make([]*db_common.Day, 0, len(m.Days))
			for i := range m.Days {
				days = append(days, m.Days[i].ToCommon(m.Year, m.Month))
			}
			return days
		}(),
	}
}

// Day represents day entity in PostgreSQL db
type Day struct {
	tableName struct{} `pg:"days"` // nolint:structcheck,unused

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	ID uint `pg:"id,pk"`

	Day int `pg:"day"`
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  money.Money `pg:"saldo,use_zero"`
	Spends []*Spend    `pg:"fk:day_id"`
}

// ToCommon converts Day to common Day structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (d *Day) ToCommon(year int, month time.Month) *db_common.Day {
	if d == nil {
		return nil
	}
	return &db_common.Day{
		ID:    d.ID,
		Year:  year,
		Month: month,
		Day:   d.Day,
		Saldo: d.Saldo,
		Spends: func() []*db_common.Spend {
			spends := make([]*db_common.Spend, 0, len(d.Spends))
			for i := range d.Spends {
				spends = append(spends, d.Spends[i].ToCommon(year, month, d.Day))
			}
			return spends
		}(),
	}
}

// Income represents income entity in PostgreSQL db
type Income struct {
	tableName struct{} `pg:"incomes"` // nolint:structcheck,unused

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	ID uint `pg:"id,pk"`

	Title  string      `pg:"title"`
	Notes  string      `pg:"notes"`
	Income money.Money `pg:"income"`
}

// Check checks whether Income is valid (not empty title, positive income and etc.)
func (in Income) Check() error {
	// Check Title
	if in.Title == "" {
		return errors.New("title can't be empty")
	}

	// Skip Notes

	// Check Income
	if in.Income <= 0 {
		return fmt.Errorf("invalid income: '%d'", in.Income)
	}

	return nil
}

// ToCommon converts Income to common Income structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (in *Income) ToCommon(year int, month time.Month) *db_common.Income {
	if in == nil {
		return nil
	}
	return &db_common.Income{
		ID:     in.ID,
		Year:   year,
		Month:  month,
		Title:  in.Title,
		Notes:  in.Notes,
		Income: in.Income,
	}
}

// MonthlyPayment represents monthly payment entity in PostgreSQL db
type MonthlyPayment struct {
	tableName struct{} `pg:"monthly_payments"` // nolint:structcheck,unused

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	ID uint `pg:"id,pk"`

	Title  string      `pg:"title"`
	TypeID uint        `pg:"type_id"`
	Type   *SpendType  `pg:"fk:type_id"`
	Notes  string      `pg:"notes"`
	Cost   money.Money `pg:"cost"`
}

// Check checks whether Monthly Payment is valid (not empty title, positive cost and etc.)
func (mp MonthlyPayment) Check() error {
	// Check Title
	if mp.Title == "" {
		return errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if mp.Cost <= 0 {
		return fmt.Errorf("invalid cost: '%d'", mp.Cost)
	}

	return nil
}

// ToCommon converts MonthlyPayment to common MonthlyPayment structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (mp *MonthlyPayment) ToCommon(year int, month time.Month) *db_common.MonthlyPayment {
	if mp == nil {
		return nil
	}
	return &db_common.MonthlyPayment{
		ID:    mp.ID,
		Year:  year,
		Month: month,
		Title: mp.Title,
		Type:  mp.Type.ToCommon(),
		Notes: mp.Notes,
		Cost:  mp.Cost,
	}
}

// Spend represents spend entity in PostgreSQL db
type Spend struct {
	tableName struct{} `pg:"spends"` // nolint:structcheck,unused

	// DayID is a foreign key to 'days' table
	DayID uint `pg:"day_id"`

	ID uint `pg:"id,pk"`

	Title  string      `pg:"title"`
	TypeID uint        `pg:"type_id"`
	Type   *SpendType  `pg:"fk:type_id"`
	Notes  string      `pg:"notes"`
	Cost   money.Money `pg:"cost"`
}

// Check checks whether Spend is valid (not empty title, positive cost and etc.)
func (s Spend) Check() error {
	// Check Title
	if s.Title == "" {
		return errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if s.Cost <= 0 {
		return fmt.Errorf("invalid cost: '%d'", s.Cost)
	}

	return nil
}

// ToCommon converts Spend to common Spend structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (s *Spend) ToCommon(year int, month time.Month, day int) *db_common.Spend {
	if s == nil {
		return nil
	}
	return &db_common.Spend{
		ID:    s.ID,
		Year:  year,
		Month: month,
		Day:   day,
		Title: s.Title,
		Type:  s.Type.ToCommon(),
		Notes: s.Notes,
		Cost:  s.Cost,
	}
}

// SpendType represents spend type entity in PostgreSQL db
type SpendType struct {
	tableName struct{} `pg:"spend_types"` // nolint:structcheck,unused

	ID   uint   `pg:"id,pk"`
	Name string `pg:"name"`
}

// Check checks whether Spend Type is valid (not empty name)
func (s SpendType) Check() error {
	// Check Name
	if s.Name == "" {
		return fmt.Errorf("name can't be empty")
	}

	return nil
}

// ToCommon converts SpendType to common SpendType structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (s *SpendType) ToCommon() *db_common.SpendType {
	if s == nil {
		return nil
	}
	return &db_common.SpendType{
		ID:   s.ID,
		Name: s.Name,
	}
}
