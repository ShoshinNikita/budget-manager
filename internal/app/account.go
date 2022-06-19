package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type AccountStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
	Get(ctx context.Context) ([]Account, error)
	Create(ctx context.Context, acc Account) error
	Update(ctx context.Context, acc Account) error
}

type Account struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Currency  money.Currency `json:"currency"`
	Status    AccountStatus  `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type AccountStatus string

const (
	AccountStatusOpen   AccountStatus = "open"
	AccountStatusClosed AccountStatus = "closed"
)

func (acc Account) GetID() uuid.UUID {
	return acc.ID
}

func (Account) GetEntityName() string {
	return "account"
}
