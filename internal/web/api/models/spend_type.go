package models

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
)

type GetSpendTypesResp struct {
	Response

	SpendTypes []db.SpendType `json:"spend_types"`
}

type AddSpendTypeReq struct {
	Request

	Name     string `json:"name" validate:"required" example:"Food"`
	ParentID uint   `json:"parent_id"`
}

func (req AddSpendTypeReq) Check() error {
	if req.Name == "" {
		return emptyFieldError("name")
	}
	return nil
}

type AddSpendTypeResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendTypeReq struct {
	Request

	ID       uint    `json:"id" validate:"required" example:"1"`
	Name     *string `json:"name" example:"Vegetables"`
	ParentID *uint   `json:"parent_id" example:"1"`
}

func (req EditSpendTypeReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	if req.Name != nil && *req.Name == "" {
		return emptyFieldError("name")
	}
	return nil
}

type RemoveSpendTypeReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req RemoveSpendTypeReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}
