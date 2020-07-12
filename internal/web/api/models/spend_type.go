package models

import (
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

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
