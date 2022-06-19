package app

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CategoryStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, category Category) error
	Update(ctx context.Context, category Category) error
}

type Category struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  uuid.UUID  `json:"parent_id"`
	Name      string     `json:"name"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (category Category) GetID() uuid.UUID {
	return category.ID
}

func (Category) GetEntityName() string {
	return "category"
}

func (category Category) IsDeleted() bool {
	return category.DeletedAt != nil
}
