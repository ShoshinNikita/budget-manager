package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/store"
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
)

const bucketName = "transactions"

type Bolt struct {
	base *store.BaseBolt[transactions.Transaction]
}

var _ transactions.Store = (*Bolt)(nil)

func NewBolt(boltStore *bbolt.DB) (*Bolt, error) {
	store := &Bolt{
		base: store.NewBaseBolt(
			boltStore, bucketName, marshalBoltTransaction, unmarshalBoltTransaction,
		),
	}

	if err := store.base.Init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (bolt Bolt) Get(ctx context.Context, args transactions.GetTransactionsArgs) ([]transactions.Transaction, error) {
	return bolt.base.GetAll(
		func(t transactions.Transaction) bool {
			if !args.IncludeDeleted && t.IsDeleted() {
				return false
			}
			return true
		},
		func(transactions []transactions.Transaction) {
			sort.Slice(transactions, func(i, j int) bool {
				return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
			})
		},
	)
}

func (bolt Bolt) GetByID(ctx context.Context, id uuid.UUID) (transactions.Transaction, error) {
	return bolt.base.GetByID(id)
}

func (bolt Bolt) Create(ctx context.Context, transactions ...transactions.Transaction) error {
	return bolt.base.Create(transactions...)
}

func (bolt Bolt) Update(ctx context.Context, transaction transactions.Transaction) error {
	return bolt.base.Update(transaction)
}

type boltTransaction struct {
	ID          uuid.UUID                    `json:"id"`
	AccountID   uuid.UUID                    `json:"account_id"`
	Type        transactions.TransactionType `json:"type"`
	Flags       transactions.TransactionFlag `json:"flags"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Amount      money.Money                  `json:"amount"`
	Extra       json.RawMessage              `json:"extra"`
	CategoryID  uuid.UUID                    `json:"category_id"`
	CreatedAt   time.Time                    `json:"created_at"`
	DeletedAt   *time.Time                   `json:"deleted_at"`
}

func marshalBoltTransaction(t transactions.Transaction) []byte {
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

func unmarshalBoltTransaction(data []byte) (transactions.Transaction, error) {
	var t boltTransaction
	if err := json.Unmarshal(data, &t); err != nil {
		return transactions.Transaction{}, errors.Wrap(err, "couldn't unmarshal transaction")
	}
	extra, err := transactions.UnmarshalTransactionExtra(t.Extra, t.Flags)
	if err != nil {
		return transactions.Transaction{}, errors.Wrap(err, "couldn't unmarshal extra")
	}

	return transactions.Transaction{
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
