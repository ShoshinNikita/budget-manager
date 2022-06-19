package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type TransactionsStore struct {
	base *BaseStore[app.Transaction]
}

func NewTransactionsStore(boltStore *bbolt.DB) (*TransactionsStore, error) {
	store := &TransactionsStore{
		base: NewBaseStore(
			boltStore, "transactions", marshalBoltTransaction, unmarshalBoltTransaction,
		),
	}

	if err := store.base.Init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (store TransactionsStore) Get(ctx context.Context, args app.GetTransactionsArgs) ([]app.Transaction, error) {
	categoryIDs := make(map[uuid.UUID]bool, len(args.CategoryIDs))
	for _, id := range args.CategoryIDs {
		categoryIDs[id] = true
	}

	return store.base.GetAll(
		func(t app.Transaction) bool {
			if !args.IncludeDeleted && t.IsDeleted() {
				return true
			}
			if len(categoryIDs) > 0 && !categoryIDs[t.CategoryID] {
				return true
			}
			return false
		},
		func(transactions []app.Transaction) {
			sort.Slice(transactions, func(i, j int) bool {
				return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
			})
		},
	)
}

func (store TransactionsStore) GetByID(ctx context.Context, id uuid.UUID) (app.Transaction, error) {
	return store.base.GetByID(id)
}

func (store TransactionsStore) Create(ctx context.Context, transactions ...app.Transaction) error {
	return store.base.Create(transactions...)
}

func (store TransactionsStore) Update(ctx context.Context, transaction app.Transaction) error {
	return store.base.Update(transaction)
}

type boltTransaction struct {
	ID          uuid.UUID           `json:"id"`
	AccountID   uuid.UUID           `json:"account_id"`
	Type        app.TransactionType `json:"type"`
	Flags       app.TransactionFlag `json:"flags"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Amount      money.Money         `json:"amount"`
	Extra       json.RawMessage     `json:"extra"`
	CategoryID  uuid.UUID           `json:"category_id"`
	CreatedAt   time.Time           `json:"created_at"`
	DeletedAt   *time.Time          `json:"deleted_at"`
}

func marshalBoltTransaction(t app.Transaction) []byte {
	rawExtra, err := json.Marshal(t.Extra)
	if err != nil {
		panic(fmt.Sprintf("error during transaction extra marshal: %s", err))
	}

	data, err := json.Marshal(boltTransaction{
		ID:          t.ID,
		AccountID:   t.AccountID,
		Type:        t.Type,
		Flags:       t.Flags,
		Name:        t.Name,
		Description: t.Description,
		Amount:      t.Amount,
		Extra:       rawExtra,
		CategoryID:  t.CategoryID,
		CreatedAt:   t.CreatedAt,
		DeletedAt:   t.DeletedAt,
	})
	if err != nil {
		panic(fmt.Sprintf("error during transaction marshal: %s", err))
	}
	return data
}

func unmarshalBoltTransaction(data []byte) (app.Transaction, error) {
	var t boltTransaction
	if err := json.Unmarshal(data, &t); err != nil {
		return app.Transaction{}, errors.Wrap(err, "couldn't unmarshal transaction")
	}
	extra, err := app.UnmarshalTransactionExtra(t.Extra, t.Flags)
	if err != nil {
		return app.Transaction{}, errors.Wrap(err, "couldn't unmarshal extra")
	}

	return app.Transaction{
		ID:          t.ID,
		AccountID:   t.AccountID,
		Type:        t.Type,
		Flags:       t.Flags,
		Name:        t.Name,
		Description: t.Description,
		Amount:      t.Amount,
		Extra:       extra,
		CategoryID:  t.CategoryID,
		CreatedAt:   t.CreatedAt,
		DeletedAt:   t.DeletedAt,
	}, nil
}
