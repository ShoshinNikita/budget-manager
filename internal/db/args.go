package db

import (
	"time"

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

// ----------------------------------------------------
// Other
// ----------------------------------------------------

// SearchSpendsArgs is used to search for spends. All fields are optional
type SearchSpendsArgs struct {
	Title string // Must be in lovercase
	Notes string // Must be in lovercase

	// TitleExactly defines should we search exactly for the given title
	TitleExactly bool
	// NotesExactly defines should we search exactly for the given notes
	NotesExactly bool

	After  time.Time
	Before time.Time

	MinCost money.Money
	MaxCost money.Money

	TypeIDs []uint
}
