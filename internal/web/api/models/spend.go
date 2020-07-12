package models

import (
	"github.com/pkg/errors"
)

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
