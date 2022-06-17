package categories

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	GetAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, name string, parentID uuid.UUID) (Category, error)
	Update(ctx context.Context, category Category) error
}

type Store interface {
	GetAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, category Category) error
	Update(ctx context.Context, category Category) error
}

type Category struct {
	ID       uuid.UUID `json:"id"`
	ParentID uuid.UUID `json:"parent_id"`
	Name     string    `json:"name"`
}

func (category Category) GetID() uuid.UUID {
	return category.ID
}
