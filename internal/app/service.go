package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type Service interface {
	GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error)
	GetAccounts(ctx context.Context) ([]Account, error)
	CreateAccount(ctx context.Context, name string, currency money.Currency) (Account, error)
	CloseAccount(ctx context.Context, id uuid.UUID) error

	GetTransactions(ctx context.Context, args GetTransactionsArgs) ([]Transaction, error)
	CalculateAccountBalances(ctx context.Context, accountIDs []uuid.UUID) (map[uuid.UUID]money.Money, error)
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (Transaction, error)
	CreateTransferTransactions(ctx context.Context, args CreateTransferTransactionsArgs) ([2]Transaction, error)
	DeleteTransaction(ctx context.Context, id uuid.UUID) error

	GetCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, name string, parentID uuid.UUID) (Category, error)
	UpdateCategory(ctx context.Context, category Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
