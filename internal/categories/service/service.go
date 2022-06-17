package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/categories"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type Service struct {
	store categories.Store
}

var _ categories.Service = (*Service)(nil)

func NewService(store categories.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) GetAll(ctx context.Context) ([]categories.Category, error) {
	return s.store.GetAll(ctx)
}

func (s Service) Create(ctx context.Context, name string, parentID uuid.UUID) (categories.Category, error) {
	category := categories.Category{
		ID:       uuid.New(),
		ParentID: parentID,
		Name:     name,
	}

	if err := s.store.Create(ctx, category); err != nil {
		return categories.Category{}, errors.Wrap(err, "couldn't save new category")
	}
	return category, nil
}

func (s Service) Update(ctx context.Context, category categories.Category) error {
	return s.store.Update(ctx, category)
}
