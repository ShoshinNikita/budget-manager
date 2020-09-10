package db

import (
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type Month struct {
	ID uint `json:"id"`

	Year  int        `json:"year"`
	Month time.Month `json:"month" swaggertype:"integer"`

	Incomes         []Income         `json:"incomes"`
	MonthlyPayments []MonthlyPayment `json:"monthly_payments"`

	// DailyBudget is a (TotalIncome - Cost of Monthly Payments) / Number of Days
	DailyBudget money.Money `json:"daily_budget" swaggertype:"number"`
	Days        []Day       `json:"days"`

	TotalIncome money.Money `json:"total_income" swaggertype:"number"`
	// TotalSpend is a cost of all Monthly Payments and Spends
	TotalSpend money.Money `json:"total_spend" swaggertype:"number"`
	// Result is TotalIncome - TotalSpend
	Result money.Money `json:"result" swaggertype:"number"`
}

type Day struct {
	ID uint `json:"id"`

	Year  int        `json:"year"`
	Month time.Month `json:"month" swaggertype:"integer"`

	Day int `json:"day"`
	// Saldo is DailyBudget - Cost of all Spends. It can be negative
	Saldo  money.Money `json:"saldo" swaggertype:"number"`
	Spends []Spend     `json:"spends"`
}

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	ID uint `json:"id"`

	Year  int        `json:"year"`
	Month time.Month `json:"month" swaggertype:"integer"`

	Title  string      `json:"title"`
	Notes  string      `json:"notes,omitempty"`
	Income money.Money `json:"income" swaggertype:"number"`
}

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	ID uint `json:"id"`

	Year  int        `json:"year"`
	Month time.Month `json:"month" swaggertype:"integer"`

	Title string      `json:"title"`
	Type  *SpendType  `json:"type,omitempty"`
	Notes string      `json:"notes,omitempty"`
	Cost  money.Money `json:"cost" swaggertype:"number"`
}

// Spend contains information about spends
type Spend struct {
	ID uint `json:"id"`

	Year  int        `json:"year"`
	Month time.Month `json:"month" swaggertype:"integer"`
	Day   int        `json:"day"`

	Title string      `json:"title"`
	Type  *SpendType  `json:"type,omitempty"`
	Notes string      `json:"notes,omitempty"`
	Cost  money.Money `json:"cost" swaggertype:"number"`
}

// SpendType contains information about spend type
type SpendType struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	ParentID uint   `json:"parent_id"`
}
