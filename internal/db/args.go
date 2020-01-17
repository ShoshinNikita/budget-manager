package db

import (
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// ----------------------------------------------------
// Income
// ----------------------------------------------------

type AddIncomeArgs struct {
	MonthID uint
	Title   string
	Notes   string
	Income  money.Money
}

type EditIncomeArgs struct {
	ID     uint
	Title  *string
	Notes  *string
	Income *money.Money
}

// ----------------------------------------------------
// Monthly Payment
// ----------------------------------------------------

type AddMonthlyPaymentArgs struct {
	MonthID uint
	Title   string
	TypeID  uint
	Notes   string
	Cost    money.Money
}

type EditMonthlyPaymentArgs struct {
	ID uint

	Title  *string
	TypeID *uint
	Notes  *string
	Cost   *money.Money
}

// ----------------------------------------------------
// Spend
// ----------------------------------------------------

type AddSpendArgs struct {
	DayID  uint
	Title  string
	TypeID uint   // optional
	Notes  string // optional
	Cost   money.Money
}

type EditSpendArgs struct {
	ID     uint
	Title  *string
	TypeID *uint
	Notes  *string
	Cost   *money.Money
}
