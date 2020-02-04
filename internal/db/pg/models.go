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
	ID uint

	Year  int
	Month time.Month

	Incomes         []*Income `pg:"fk:month_id"`
	TotalIncome     money.Money
	DailyBudget     money.Money
	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id"`
	Days            []*Day            `pg:"fk:month_id"`
	TotalSpend      money.Money
	Result          money.Money
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
				incomes = append(incomes, m.Incomes[i].ToCommon())
			}
			return incomes
		}(),
		MonthlyPayments: func() []*db_common.MonthlyPayment {
			mp := make([]*db_common.MonthlyPayment, 0, len(m.MonthlyPayments))
			for i := range m.MonthlyPayments {
				mp = append(mp, m.MonthlyPayments[i].ToCommon())
			}
			return mp
		}(),
		Days: func() []*db_common.Day {
			days := make([]*db_common.Day, 0, len(m.Days))
			for i := range m.Days {
				days = append(days, m.Days[i].ToCommon())
			}
			return days
		}(),
	}
}

// Day represents day entity in PostgreSQL db
type Day struct {
	// MonthID is a foreign key to 'months' table
	MonthID uint

	ID uint

	Day    int
	Saldo  money.Money
	Spends []*Spend `pg:"fk:day_id"`
}

// ToCommon converts Day to common Day structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (d *Day) ToCommon() *db_common.Day {
	if d == nil {
		return nil
	}
	return &db_common.Day{
		MonthID: d.MonthID,
		ID:      d.ID,
		Day:     d.Day,
		Saldo:   d.Saldo,
		Spends: func() []*db_common.Spend {
			spends := make([]*db_common.Spend, 0, len(d.Spends))
			for i := range d.Spends {
				spends = append(spends, d.Spends[i].ToCommon())
			}
			return spends
		}(),
	}
}

// Income represents income entity in PostgreSQL db
type Income struct {
	// MonthID is a foreign key to 'months' table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	Notes  string
	Income money.Money
}

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
func (in *Income) ToCommon() *db_common.Income {
	if in == nil {
		return nil
	}
	return &db_common.Income{
		MonthID: in.MonthID,
		ID:      in.ID,
		Title:   in.Title,
		Notes:   in.Notes,
		Income:  in.Income,
	}
}

// MonthlyPayment represents monthly payment entity in PostgreSQL db
type MonthlyPayment struct {
	// MonthID is a foreign key to 'months' table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   money.Money
}

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
func (mp *MonthlyPayment) ToCommon() *db_common.MonthlyPayment {
	if mp == nil {
		return nil
	}
	return &db_common.MonthlyPayment{
		MonthID: mp.MonthID,
		ID:      mp.ID,
		Title:   mp.Title,
		TypeID:  mp.TypeID,
		Type:    mp.Type.ToCommon(),
		Notes:   mp.Notes,
		Cost:    mp.Cost,
	}
}

// Spend represents spend entity in PostgreSQL db
type Spend struct {
	// DayID is a foreign key to 'days' table
	DayID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   money.Money
}

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
func (s *Spend) ToCommon() *db_common.Spend {
	if s == nil {
		return nil
	}
	return &db_common.Spend{
		DayID:  s.DayID,
		ID:     s.ID,
		Title:  s.Title,
		TypeID: s.TypeID,
		Type:   s.Type.ToCommon(),
		Notes:  s.Notes,
		Cost:   s.Cost,
	}
}

// SpendType represents spend type entity in PostgreSQL db
type SpendType struct {
	ID   uint `pg:",pk"`
	Name string
}

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
