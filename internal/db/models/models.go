package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/ShoshinNikita/budget_manager/internal/pkg/money"
)

// ----------------------------------------------------
// Common
// ----------------------------------------------------

type Month struct {
	ID    uint       `json:"id"`
	Year  int        `json:"year"`
	Month time.Month `json:"month"`

	// Incomes

	Incomes     []*Income   `pg:"fk:month_id" json:"incomes"`
	TotalIncome money.Money `json:"total_income"`
	// DailyBudget is a (TotalIncome - Cost of Monthly Payments) / Number of Days
	DailyBudget money.Money `json:"daily_budget"`

	// Spends

	MonthlyPayments []*MonthlyPayment `pg:"fk:month_id" json:"monthly_payments"`
	Days            []*Day            `pg:"fk:month_id" json:"days"`
	TotalSpend      money.Money       `json:"total_spend"` // must be negative or zero

	// Result is TotalIncome - TotalSpend
	Result money.Money `json:"result"`
}

type Day struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint `json:"month_id"`

	ID uint `json:"id"`

	Day int `json:"day"`
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  money.Money `json:"saldo"`
	Spends []*Spend    `pg:"fk:day_id" json:"spends"`
}

// ----------------------------------------------------
// Income
// ----------------------------------------------------

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	// MonthID is a foreign key to Months table
	MonthID uint `json:"month_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	Notes  string      `json:"notes,omitempty"`
	Income money.Money `json:"income"`
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

// ----------------------------------------------------
// Monthly Payment
// ----------------------------------------------------

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint `json:"month_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	TypeID uint        `json:"type_id,omitempty"`
	Type   *SpendType  `pg:"fk:type_id" json:"type,omitempty"`
	Notes  string      `json:"notes,omitempty"`
	Cost   money.Money `json:"cost"`
}

// Check checks whether Monthly Payment is valid (not empty title, positive cost and etc.)
func (in MonthlyPayment) Check() error {
	// Check Title
	if in.Title == "" {
		return errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if in.Cost <= 0 {
		return fmt.Errorf("invalid cost: '%d'", in.Cost)
	}

	return nil
}

// ----------------------------------------------------
// Spend
// ----------------------------------------------------

// Spend contains information about spends
type Spend struct {
	// MonthID is a foreign key to Days table
	DayID uint `json:"day_id"`

	ID uint `pg:",pk" json:"-"`

	Title  string      `json:"title"`
	TypeID uint        `json:"type_id,omitempty"`
	Type   *SpendType  `pg:"fk:type_id" json:"type,omitempty"`
	Notes  string      `json:"notes,omitempty"`
	Cost   money.Money `json:"cost"`
}

// Check checks whether Income is valid (not empty title, positive cost and etc.)
func (in Spend) Check() error {
	// Check Title
	if in.Title == "" {
		return errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if in.Cost <= 0 {
		return fmt.Errorf("invalid cost: '%d'", in.Cost)
	}

	return nil
}

// ----------------------------------------------------
// Spend Type
// ----------------------------------------------------

// SpendType contains information about spend type
type SpendType struct {
	ID   uint   `pg:",pk" json:"id"`
	Name string `json:"name"`
}

// Check checks whether Spend Type is valid (not empty name)
func (in SpendType) Check() error {
	// Check Name
	if in.Name == "" {
		return fmt.Errorf("name can't be empty")
	}

	return nil
}
