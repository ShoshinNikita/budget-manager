// Package models contains models of requests and responses
package models

import (
	"errors"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// All requests implement the Request interface
type Request interface {
	request()
}

// BaseRequest is a base request model that implements Request interface.
// It must be nested into all requests
type BaseRequest struct{}

func (BaseRequest) request() {}

// BaseResponse is a base response model that must be nested into all responses
type BaseResponse struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	// Error is specified only when success if false
	Error string `json:"error,omitempty"`
}

// -------------------------------------------------
// Month
// -------------------------------------------------

type GetMonthByDateReq struct {
	BaseRequest

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
	BaseResponse

	Month db.Month `json:"month"`
}
