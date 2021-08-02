// Package models contains models of requests and responses
package models

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// Request is a base request model that must be nested into all requests
type Request struct{}

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

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req *GetMonthByIDReq) SanitizeAndCheck() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}

type GetMonthByDateReq struct {
	Request

	Year  int `json:"year" validate:"required" example:"2020"`
	Month int `json:"month" validate:"required" example:"7"`
}

func (req *GetMonthByDateReq) SanitizeAndCheck() error {
	if req.Year == 0 {
		return emptyOrZeroFieldError("year")
	}
	if req.Month == 0 {
		return emptyOrZeroFieldError("month")
	}
	return nil
}

type GetMonthResp struct {
	Response

	Month db.Month `json:"month"`
}
