package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type Service struct {
	store app.CategoryStore
}

func NewService(store app.CategoryStore) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) GetAll(ctx context.Context) ([]app.Category, error) {
	return s.store.GetAll(ctx)
}

func (s Service) Create(ctx context.Context, name string, parentID uuid.UUID) (app.Category, error) {
	category := app.Category{
		ID:       uuid.New(),
		ParentID: parentID,
		Name:     name,
	}

	if err := s.store.Create(ctx, category); err != nil {
		return app.Category{}, errors.Wrap(err, "couldn't save new category")
	}
	return category, nil
}

func (s Service) Update(ctx context.Context, category app.Category) error {
	return s.store.Update(ctx, category)
}
