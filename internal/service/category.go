package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

func (s Service) GetCategoryByID(ctx context.Context, id uuid.UUID) (app.Category, error) {
	return s.categoryStore.GetByID(ctx, id)
}

func (s Service) GetCategories(ctx context.Context) ([]app.Category, error) {
	return s.categoryStore.GetAll(ctx)
}

func (s Service) CreateCategory(ctx context.Context, name string, parentID uuid.UUID) (app.Category, error) {
	category := app.Category{
		ID:       uuid.New(),
		ParentID: parentID,
		Name:     name,
	}

	if err := s.categoryStore.Create(ctx, category); err != nil {
		return app.Category{}, errors.Wrap(err, "couldn't save new category")
	}
	return category, nil
}

func (s Service) UpdateCategory(ctx context.Context, category app.Category) error {
	return s.categoryStore.Update(ctx, category)
}
