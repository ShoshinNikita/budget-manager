// Package models contains models of requests and responses
package models

// -------------------------------------------------
// Common
// -------------------------------------------------

type Request struct {
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// -------------------------------------------------
// Income
// -------------------------------------------------

type AddIncomeReq struct {
	Request

	MonthID uint   `json:"month_id"`
	Title   string `json:"title"`
	Notes   string `json:"notes,omitempty"`
	Income  int64  `json:"income"`
}
type AddIncomeResp struct {
	Response

	ID uint `json:"id"`
}

type EditIncomeReq struct {
	Request

	ID     uint    `json:"id"`
	Title  *string `json:"title,omitempty"`
	Notes  *string `json:"notes,omitempty"`
	Income *int64  `json:"income,omitempty"`
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

	Title  string `json:"title"`
	TypeID uint   `json:"type_id,omitempty"`
	Notes  string `json:"notes,omitempty"`
	Cost   int64  `json:"cost"`
}
type AddMonthlyPaymentResp struct {
	Response

	ID uint `json:"id"`
}

type EditMonthlyPaymentReq struct {
	Request

	ID     uint    `json:"id"`
	Title  *string `json:"title,omitempty"`
	TypeID *uint   `json:"type_id,omitempty"`
	Notes  *string `json:"notes,omitempty"`
	Cost   *int64  `json:"cost,omitempty"`
}

type RemoveMonthlyPaymentReq struct {
	Request

	ID uint `json:"id"`
}
