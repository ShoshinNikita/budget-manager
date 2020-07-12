package models

import "github.com/pkg/errors"

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
