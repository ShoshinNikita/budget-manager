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
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
)

const bucketName = "transactions"

type Bolt struct {
	store *bbolt.DB
}

var _ transactions.Store = (*Bolt)(nil)

func NewBolt(boltStore *bbolt.DB) (*Bolt, error) {
	store := &Bolt{
		store: boltStore,
	}

	if err := store.init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (s Bolt) init() error {
	return s.store.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return errors.Wrapf(err, "couldn't create bucket %q", bucketName)
	})
}

func (s Bolt) Get(ctx context.Context, args transactions.GetTransactionsArgs) ([]transactions.Transaction, error) {
	var res []transactions.Transaction

	err := s.store.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		return b.ForEach(func(k, v []byte) error {
			t, err := unmarshalBoltTransaction(v)
			if err != nil {
				return err
			}

			// TODO: apply filters

			res = append(res, t)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.Before(res[j].CreatedAt)
	})

	return res, nil
}

func (s Bolt) Create(ctx context.Context, transactions ...transactions.Transaction) error {
	return s.store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		for _, t := range transactions {
			transactionData := marshalBoltTransaction(t)
			if err := b.Put(t.ID[:], transactionData); err != nil {
				return errors.Wrapf(err, "couldn't put transaction %#v", t)
			}
		}
		return nil
	})
}

type boltTransaction struct {
	ID         uuid.UUID                     `json:"id"`
	AccountID  uuid.UUID                     `json:"account_id"`
	Type       transactions.TransactionType  `json:"type"`
	Flags      transactions.TransactionFlag  `json:"flags"`
	Amount     money.Money                   `json:"amount"`
	Extra      transactions.TransactionExtra `json:"extra,omitempty"`
	CategoryID uuid.UUID                     `json:"category_id"`
	CreatedAt  time.Time                     `json:"created_at"`
}

func marshalBoltTransaction(t transactions.Transaction) []byte {
	data, err := json.Marshal(boltTransaction{
		ID:         t.ID,
		AccountID:  t.AccountID,
		Type:       t.Type,
		Flags:      t.Flags,
		Amount:     t.Amount,
		Extra:      t.Extra,
		CategoryID: t.CategoryID,
		CreatedAt:  t.CreatedAt,
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

	return transactions.Transaction{
		ID:         t.ID,
		AccountID:  t.AccountID,
		Type:       t.Type,
		Flags:      t.Flags,
		Amount:     t.Amount,
		Extra:      t.Extra,
		CategoryID: t.CategoryID,
		CreatedAt:  t.CreatedAt,
	}, nil
}
