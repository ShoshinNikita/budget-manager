package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type TransactionStore interface {
	Get(ctx context.Context, args GetTransactionsArgs) ([]Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (Transaction, error)
	Create(ctx context.Context, transactions ...Transaction) error
	Update(ctx context.Context, transaction Transaction) error
}

// TODO: add filters and limit, offset
type GetTransactionsArgs struct {
	IncludeDeleted bool
	CategoryIDs    []uuid.UUID
}

type CreateTransactionArgs struct {
	AccountID   uuid.UUID
	Type        TransactionType
	Name        string
	Description string
	Amount      money.Money
	CategoryID  uuid.UUID
}

type CreateTransferTransactionsArgs struct {
	FromAccountID uuid.UUID
	FromAmount    money.Money

	ToAccountID uuid.UUID
	ToAmount    money.Money
}

type Transaction struct {
	ID          uuid.UUID        `json:"id"`
	AccountID   uuid.UUID        `json:"account_id"`
	Type        TransactionType  `json:"type"`
	Flags       TransactionFlag  `json:"flags"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Amount      money.Money      `json:"amount"`
	Extra       TransactionExtra `json:"extra,omitempty"`
	CategoryID  uuid.UUID        `json:"category_id"`
	CreatedAt   time.Time        `json:"created_at"`
	DeletedAt   *time.Time       `json:"deleted_at"`
}

type TransactionType string

const (
	TransactionTypeAdd      TransactionType = "add"
	TransactionTypeWithdraw TransactionType = "withdraw"
)

func (t TransactionType) IsValid() error {
	if t == TransactionTypeAdd || t == TransactionTypeWithdraw {
		return nil
	}
	return errors.Errorf("invalid transaction type %q", t)
}

type TransactionFlag int

const (
	TransactionFlagTransfer TransactionFlag = 1 << iota
)

func (f TransactionFlag) IsTransferTransaction() bool {
	return f&TransactionFlagTransfer != 0
}

type TransactionExtra interface {
	extra()
}

type TransferTransactionExtra struct {
	TransferID uuid.UUID `json:"transfer_id"`
}

func (*TransferTransactionExtra) extra() {}

func UnmarshalTransactionExtra(data []byte, flag TransactionFlag) (TransactionExtra, error) {
	var extra TransactionExtra

	switch { //nolint:gocritic
	case flag.IsTransferTransaction():
		extra = &TransferTransactionExtra{}
	}
	if extra != nil {
		if err := json.Unmarshal(data, &extra); err != nil {
			return nil, errors.Wrap(err, "couldn't unmarshal extra")
		}
	}

	return extra, nil
}

func (t Transaction) GetID() uuid.UUID {
	return t.ID
}

func (Transaction) GetEntityName() string {
	return "transaction"
}

func (t Transaction) IsDeleted() bool {
	return t.DeletedAt != nil
}
