package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/store"
)

func (s Service) GetCategoryByID(ctx context.Context, id uuid.UUID) (app.Category, error) {
	category, err := s.categoryStore.GetByID(ctx, id)
	if err != nil {
		return app.Category{}, err
	}
	if category.IsDeleted() {
		return app.Category{}, &store.NotFoundError{
			EntityName: category.GetEntityName(),
			ID:         id,
		}
	}
	return category, nil
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

func (s Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	category, err := s.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}

	transactions, err := s.GetTransactions(ctx, app.GetTransactionsArgs{
		CategoryIDs: []uuid.UUID{category.ID},
	})
	if err != nil {
		return errors.Wrap(err, "couldn't check if there are any transactions with category to delete")
	}
	if len(transactions) > 0 {
		return errors.Errorf("%d transactions have this category", len(transactions))
	}

	now := time.Now()
	category.DeletedAt = &now

	if err := s.categoryStore.Update(ctx, category); err != nil {
		return errors.Wrap(err, "couldn't update category for deletion")
	}
	return nil
}
