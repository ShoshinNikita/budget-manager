// Package models contains models of requests and responses
package models

import (
	"errors"
	"time"

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

type GetMonthByDateReq struct {
	Request

	Year  int        `json:"year" validate:"required" example:"2020"`
	Month time.Month `json:"month" validate:"required" swaggertype:"integer" example:"7"`
}

func (req *GetMonthByDateReq) SanitizeAndCheck() error {
	if req.Year == 0 {
		return emptyOrZeroFieldError("year")
	}
	if !(time.January <= req.Month && req.Month <= time.December) {
		return errors.New("invalid month")
	}
	return nil
}

type GetMonthResp struct {
	Response

	Month db.Month `json:"month"`
}
