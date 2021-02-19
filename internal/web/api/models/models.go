// Package models contains models of requests and responses
package models

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// Request is a base request model that must be nested into all requests
type Request struct{}

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

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req GetMonthByIDReq) Check() error {
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

func (req GetMonthByDateReq) Check() error {
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

// -------------------------------------------------
// Day
// -------------------------------------------------

type GetDayByIDReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req GetDayByIDReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}

type GetDayByDateReq struct {
	Request

	Year  int `json:"year" validate:"required" example:"2020"`
	Month int `json:"month" validate:"required" example:"7"`
	Day   int `json:"day" validate:"required" example:"13"`
}

func (req GetDayByDateReq) Check() error {
	if req.Year == 0 {
		return emptyOrZeroFieldError("year")
	}
	if req.Month == 0 {
		return emptyOrZeroFieldError("month")
	}
	if req.Day == 0 {
		return emptyOrZeroFieldError("day")
	}
	return nil
}

type GetDayResp struct {
	Response

	Day db.Day `json:"day"`
}
