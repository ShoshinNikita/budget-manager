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

type Request struct {
}

type Response struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"` // optional
}

// -------------------------------------------------
// Month
// -------------------------------------------------

type GetMonthReq struct {
	Request

	ID *uint `json:"id"`
}

type GetMonthByYearAndMonthReq struct {
	Request

	Year  *int        `json:"year"`
	Month *time.Month `json:"month"`
}

type GetMonthResp struct {
	Response

	Month db.Month `json:"month"`
}

// -------------------------------------------------
// Day
// -------------------------------------------------

type GetDayReq struct {
	Request

	ID *uint `json:"id"`
}

type GetDayByDateReq struct {
	Request

	Year  *int        `json:"year"`
	Month *time.Month `json:"month"`
	Day   *int        `json:"day"`
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

	MonthID uint    `json:"month_id"`
	Title   string  `json:"title"`
	Notes   string  `json:"notes,omitempty"` // optional
	Income  float64 `json:"income"`
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

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`  // optional
	Notes  *string  `json:"notes,omitempty"`  // optional
	Income *float64 `json:"income,omitempty"` // optional
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

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`   // optional
	TypeID *uint    `json:"type_id,omitempty"` // optional
	Notes  *string  `json:"notes,omitempty"`   // optional
	Cost   *float64 `json:"cost,omitempty"`    // optional
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

	ID     uint     `json:"id"`
	Title  *string  `json:"title,omitempty"`   // optional
	TypeID *uint    `json:"type_id,omitempty"` // optional
	Notes  *string  `json:"notes,omitempty"`   // optional
	Cost   *float64 `json:"cost,omitempty"`    // optional
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

	ID uint `json:"id"`
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

	Name string `json:"name"`
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

	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (req EditSpendTypeReq) Check() error {
	if req.Name == "" {
		return errors.New("name can't be empty")
	}
	return nil
}

type RemoveSpendTypeReq struct {
	Request

	ID uint `json:"id"`
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
	After time.Time `json:"after,omitempty"`
	// Before must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	Before time.Time `json:"before,omitempty"`

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
