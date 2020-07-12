package models

import (
	"github.com/pkg/errors"
)

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
