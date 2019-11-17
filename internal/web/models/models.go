// Package models contains models of requests and responses
package models

// -------------------------------------------------
// Common
// -------------------------------------------------

type Request struct {
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"` // optional
}

// -------------------------------------------------
// Income
// -------------------------------------------------

type AddIncomeReq struct {
	Request

	MonthID uint    `json:"month_id"`
	Title   string  `json:"title"`
	Notes   string  `json:"notes,omitempty"` // optional
	Income  float64 `json:"income"`
}
type AddIncomeResp struct {
	Response

	ID uint `json:"id"`
}

type EditIncomeReq struct {
	Request

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`  // optional
	Notes  *string  `json:"notes,omitempty"`  // optional
	Income *float64 `json:"income,omitempty"` // optional
}

type RemoveIncomeReq struct {
	Request

	ID uint `json:"id"`
}

// -------------------------------------------------
// Monthly Payment
// -------------------------------------------------

type AddMonthlyPaymentReq struct {
	Request

	MonthID uint `json:"month_id"`

	Title  string  `json:"title"`
	TypeID uint    `json:"type_id,omitempty"` // optional
	Notes  string  `json:"notes,omitempty"`   // optional
	Cost   float64 `json:"cost"`
}
type AddMonthlyPaymentResp struct {
	Response

	ID uint `json:"id"`
}

type EditMonthlyPaymentReq struct {
	Request

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`   // optional
	TypeID *uint    `json:"type_id,omitempty"` // optional
	Notes  *string  `json:"notes,omitempty"`   // optional
	Cost   *float64 `json:"cost,omitempty"`    // optional
}

type RemoveMonthlyPaymentReq struct {
	Request

	ID uint `json:"id"`
}

// -------------------------------------------------
// Spend
// -------------------------------------------------

type AddSpendReq struct {
	Request

	DayID uint `json:"day_id"`

	Title  string  `json:"title"`
	TypeID uint    `json:"type_id,omitempty"` // optional
	Notes  string  `json:"notes,omitempty"`   // optional
	Cost   float64 `json:"cost"`
}
type AddSpendResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendReq struct {
	Request

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`   // optional
	TypeID *uint    `json:"type_id,omitempty"` // optional
	Notes  *string  `json:"notes,omitempty"`   // optional
	Cost   *float64 `json:"cost,omitempty"`    // optional
}

type RemoveSpendReq struct {
	Request

	ID uint `json:"id"`
}

// -------------------------------------------------
// Spend Type
// -------------------------------------------------

type AddSpendTypeReq struct {
	Request

	Name string `json:"name"`
}
type AddSpendTypeResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendTypeReq struct {
	Request

	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type RemoveSpendTypeReq struct {
	Request

	ID uint `json:"id"`
}
