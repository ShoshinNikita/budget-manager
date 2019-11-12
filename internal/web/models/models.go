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

	ID     uint
	Title  *string `json:"title,omitempty"`
	Notes  *string `json:"notes,omitempty"`
	Income *int64  `json:"income,omitempty"`
}

type RemoveIncomeReq struct {
	Request

	ID uint
}
