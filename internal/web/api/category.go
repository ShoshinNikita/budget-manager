package api

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
)

type getCategoriesResp struct {
	Categories []app.Category `json:"categories"`
}

func (api API) getCategories(ctx context.Context, req *emptyReq) (*getCategoriesResp, error) {
	categories, err := api.service.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &getCategoriesResp{
		Categories: categories,
	}, nil
}

type (
	createCategoryReq struct {
		Name     string    `json:"name"`
		ParentID uuid.UUID `json:"parent_id"`
	}
	createCategoryResp struct {
		NewCategory app.Category `json:"new_category"`
	}
)

func (api API) createCategory(ctx context.Context, req *createCategoryReq) (*createCategoryResp, error) {
	newCategory, err := api.service.CreateCategory(ctx, req.Name, req.ParentID)
	if err != nil {
		return nil, err
	}
	return &createCategoryResp{
		NewCategory: newCategory,
	}, nil
}

type deleteCategoryReq struct {
	ID uuid.UUID `json:"id"`
}

func (api API) deleteCategory(ctx context.Context, req *deleteCategoryReq) (*emptyResp, error) {
	err := api.service.DeleteCategory(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &emptyResp{}, nil
}
