// Package models contains models of requests and responses
package models

import (
	"time"

	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// -------------------------------------------------
// Common
// -------------------------------------------------

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

	ID uint `json:"id" validate:"required" example:"1"`
}

type GetMonthByDateReq struct {
	Request

	Year  int `json:"year" validate:"required" example:"2020"`
	Month int `json:"month" validate:"required" example:"4"`
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

type GetDayByDateReq struct {
	Request

	Year  int `json:"year" validate:"required" example:"2020"`
	Month int `json:"month" validate:"required" example:"4"`
	Day   int `json:"day" validate:"required" example:"12"`
}

type GetDayResp struct {
	Response

	Day db.Day `json:"day"`
}

// -------------------------------------------------
// Income
// -------------------------------------------------

type AddIncomeReq struct {
	Request

	MonthID uint    `json:"month_id" validate:"required" example:"1"`
	Title   string  `json:"title" validate:"required" example:"Salary"`
	Notes   string  `json:"notes,omitempty"` // optional
	Income  float64 `json:"income" validate:"required" example:"10000"`
}

func (req AddIncomeReq) Check() error {
	if req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Notes
	if req.Income <= 0 {
		return errors.Errorf("invalid income: '%.2f'", req.Income)
	}
	return nil
}

type AddIncomeResp struct {
	Response

	ID uint `json:"id"`
}

type EditIncomeReq struct {
	Request

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title,omitempty"`                     // optional
	Notes  *string  `json:"notes,omitempty" example:"New notes"` // optional
	Income *float64 `json:"income,omitempty" example:"15000"`    // optional
}

func (req EditIncomeReq) Check() error {
	if req.Title != nil && *req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Notes
	if req.Income != nil && *req.Income <= 0 {
		return errors.Errorf("invalid income: '%.2f'", *req.Income)
	}
	return nil
}

type RemoveIncomeReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

// -------------------------------------------------
// Monthly Payment
// -------------------------------------------------

type AddMonthlyPaymentReq struct {
	Request

	MonthID uint `json:"month_id" validate:"required" example:"1"`

	Title  string  `json:"title" validate:"required" example:"Rent"`
	TypeID uint    `json:"type_id,omitempty"` // optional
	Notes  string  `json:"notes,omitempty"`   // optional
	Cost   float64 `json:"cost" validate:"required" example:"1500"`
}

func (req AddMonthlyPaymentReq) Check() error {
	if req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Type
	// Skip Notes
	if req.Cost <= 0 {
		return errors.Errorf("invalid cost: '%.2f'", req.Cost)
	}
	return nil
}

type AddMonthlyPaymentResp struct {
	Response

	ID uint `json:"id"`
}

type EditMonthlyPaymentReq struct {
	Request

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title,omitempty"`                     // optional
	TypeID *uint    `json:"type_id,omitempty" example:"1"`       // optional
	Notes  *string  `json:"notes,omitempty" example:"New notes"` // optional
	Cost   *float64 `json:"cost,omitempty" example:"1550"`       // optional
}

func (req EditMonthlyPaymentReq) Check() error {
	if req.Title != nil && *req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Type
	// Skip Notes
	if req.Cost != nil && *req.Cost <= 0 {
		return errors.Errorf("invalid cost: '%.2f'", *req.Cost)
	}
	return nil
}

type RemoveMonthlyPaymentReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

// -------------------------------------------------
// Spend
// -------------------------------------------------

type AddSpendReq struct {
	Request

	DayID uint `json:"day_id" validate:"required"`

	Title  string  `json:"title" validate:"required" example:"Food"`
	TypeID uint    `json:"type_id,omitempty"` // optional
	Notes  string  `json:"notes,omitempty"`   // optional
	Cost   float64 `json:"cost" validate:"required" example:"30"`
}

func (req AddSpendReq) Check() error {
	if req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Type
	// Skip Notes
	if req.Cost <= 0 {
		return errors.Errorf("invalid cost: '%.2f'", req.Cost)
	}
	return nil
}

type AddSpendResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendReq struct {
	Request

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title,omitempty"`                      // optional
	TypeID *uint    `json:"type_id,omitempty"`                    // optional
	Notes  *string  `json:"notes,omitempty" example:"Vegetables"` // optional
	Cost   *float64 `json:"cost,omitempty" example:"30.15"`       // optional
}

func (req EditSpendReq) Check() error {
	if req.Title != nil && *req.Title == "" {
		return errors.New("title can't be empty")
	}
	// Skip Type
	// Skip Notes
	if req.Cost != nil && *req.Cost <= 0 {
		return errors.Errorf("invalid cost: '%.2f'", *req.Cost)
	}
	return nil
}

type RemoveSpendReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

// -------------------------------------------------
// Spend Type
// -------------------------------------------------

type GetSpendTypesResp struct {
	Response

	SpendTypes []*db.SpendType `json:"spend_types"`
}

type AddSpendTypeReq struct {
	Request

	Name string `json:"name" validate:"required" example:"Food"`
}

func (req AddSpendTypeReq) Check() error {
	if req.Name == "" {
		return errors.New("name can't be empty")
	}
	return nil
}

type AddSpendTypeResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendTypeReq struct {
	Request

	ID   uint   `json:"id" validate:"required" example:"1"`
	Name string `json:"name" example:"Vegetables"`
}

func (req EditSpendTypeReq) Check() error {
	if req.Name == "" {
		return errors.New("name can't be empty")
	}
	return nil
}

type RemoveSpendTypeReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

// -------------------------------------------------
// Other
// -------------------------------------------------

// SearchSpendsReq is used to search for spends, all fields are optional
//
// nolint:maligned
type SearchSpendsReq struct {
	Request

	// Title can be in any case. Search will be performed by lowercased value
	Title string `json:"title,omitempty"`
	// Notes can be in any case. Search will be performed by lowercased value
	Notes string `json:"notes,omitempty"`

	// TitleExactly defines should we search exactly for the given title
	TitleExactly bool `json:"title_exactly,omitempty"`
	// NotesExactly defines should we search exactly for the given notes
	NotesExactly bool `json:"notes_exactly,omitempty"`

	// After must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	After time.Time `json:"after,omitempty" format:"date"`
	// Before must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	Before time.Time `json:"before,omitempty" format:"date"`

	MinCost float64 `json:"min_cost,omitempty"`
	MaxCost float64 `json:"max_cost,omitempty"`

	// WithoutType is used to search for Spends without Spend Type. TypeIDs are ignored when it is true
	WithoutType bool `json:"without_type,omitempty"`
	// TypeIDs is a list of Spend Type ids to search for
	TypeIDs []uint `json:"type_ids,omitempty"`

	// Sort specify field to sort by. Available options: title, cost, date (default)
	Sort string `json:"sort,omitempty"`
	// Order specify sort order. Available options: asc (default), desc
	Order string `json:"order,omitempty"`
}

type SearchSpendsResp struct {
	Response

	Spends []*db.Spend `json:"spends"`
}
