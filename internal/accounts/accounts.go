package accounts

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type Service interface {
	GetAll(ctx context.Context) ([]AccountWithBalance, error)
	Create(ctx context.Context, currency money.Currency, initialAmount money.Money) error
	Close(ctx context.Context, id uuid.UUID) error
}

type Store interface {
	GetAll(ctx context.Context) ([]Account, error)
	Create(ctx context.Context, acc Account) error
	Update(ctx context.Context, acc Account) error
}

type Account struct {
	ID        uuid.UUID      `json:"id"`
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

func NewAccount(currency money.Currency) Account {
	now := time.Now()

	return Account{
		ID:        uuid.New(),
		Currency:  currency,
		Status:    AccountStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type AccountWithBalance struct {
	Account

	Balance money.Money `json:"balance"`
}
