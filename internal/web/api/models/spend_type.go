package models

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
)

type GetSpendTypesResp struct {
	BaseResponse

	SpendTypes []db.SpendType `json:"spend_types"`
}

type AddSpendTypeReq struct {
	BaseRequest

	Name     string `json:"name" validate:"required" example:"Food"`
	ParentID uint   `json:"parent_id"`
}

func (req *AddSpendTypeReq) SanitizeAndCheck() error {
	sanitizeString(&req.Name)

	if req.Name == "" {
		return emptyFieldError("name")
	}
	return nil
}

type AddSpendTypeResp struct {
	BaseResponse

	ID uint `json:"id"`
}

type EditSpendTypeReq struct {
	BaseRequest

	ID       uint    `json:"id" validate:"required" example:"1"`
	Name     *string `json:"name" example:"Vegetables"`
	ParentID *uint   `json:"parent_id" example:"1"`
}

func (req *EditSpendTypeReq) SanitizeAndCheck() error {
	sanitizeString(req.Name)

	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	if req.Name != nil && *req.Name == "" {
		return emptyFieldError("name")
	}
	return nil
}

type RemoveSpendTypeReq struct {
	BaseRequest

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req *RemoveSpendTypeReq) SanitizeAndCheck() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}
