// Package models contains models of requests and responses
package models

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// Request is a base request model that must be nested into all requests
type Request struct {
}

// Check is a default method to implement 'web.RequestChecker' interface
func (Request) Check() error {
	return nil
}

// Response is a base response model that must be nested into all responses
type Response struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	// Error is specified only when success if false
	Error string `json:"error,omitempty"`
}

// -------------------------------------------------
// Month
// -------------------------------------------------

type GetMonthByIDReq struct {
	Request

	ID uint `json:"id" validate:"required"`
}

type GetMonthByDateReq struct {
	Request

	Year  int `json:"year" validate:"required"`
	Month int `json:"month" validate:"required"`
}

type GetMonthResp struct {
	Response

	Month db.Month `json:"month"`
}

// -------------------------------------------------
// Day
// -------------------------------------------------

type GetDayByIDReq struct {
	Request

	ID uint `json:"id" validate:"required"`
}

type GetDayByDateReq struct {
	Request

	Year  int `json:"year" validate:"required"`
	Month int `json:"month" validate:"required"`
	Day   int `json:"day" validate:"required"`
}

type GetDayResp struct {
	Response

	Day db.Day `json:"day"`
}
