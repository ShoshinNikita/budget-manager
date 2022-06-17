package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/accounts"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

const bucketName = "accounts"

type Bolt struct {
	store *bbolt.DB
}

var _ accounts.Store = (*Bolt)(nil)

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

func (s Bolt) GetByID(ctx context.Context, id uuid.UUID) (accounts.Account, error) {
	var res accounts.Account

	err := s.store.View(func(tx *bbolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucketName))

		accountData := b.Get(id[:])
		if accountData == nil {
			return accounts.ErrAccountNotExist
		}
		res, err = unmarshalBoltAccount(accountData)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return accounts.Account{}, err
	}
	return res, nil
}

func (s Bolt) GetAll(ctx context.Context) (res []accounts.Account, err error) {
	err = s.store.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		return b.ForEach(func(k, v []byte) error {
			acc, err := unmarshalBoltAccount(v)
			if err != nil {
				return err
			}
			res = append(res, acc)
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

func (s Bolt) Create(ctx context.Context, acc accounts.Account) error {
	data := marshalBoltAccount(acc)

	return s.store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		if err := b.Put(acc.ID[:], data); err != nil {
			return errors.Wrap(err, "couldn't put marshalled account")
		}
		return nil
	})
}

func (s Bolt) Update(ctx context.Context, acc accounts.Account) error {
	data := marshalBoltAccount(acc)

	return s.store.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		prevValue := b.Get(acc.ID[:])
		if prevValue == nil {
			return accounts.ErrAccountNotExist
		}

		if err := b.Put(acc.ID[:], data); err != nil {
			return errors.Wrap(err, "couldn't put marshalled account")
		}
		return nil
	})
}

type boltAccount struct {
	ID        uuid.UUID              `json:"id"`
	Currency  money.Currency         `json:"currency"`
	Status    accounts.AccountStatus `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func marshalBoltAccount(acc accounts.Account) []byte {
	data, err := json.Marshal(boltAccount{
		ID:        acc.ID,
		Currency:  acc.Currency,
		Status:    acc.Status,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
	})
	if err != nil {
		panic(fmt.Sprintf("error during account marshal: %s", err))
	}
	return data
}

func unmarshalBoltAccount(data []byte) (accounts.Account, error) {
	var acc boltAccount
	if err := json.Unmarshal(data, &acc); err != nil {
		return accounts.Account{}, errors.Wrap(err, "couldn't unmarshal account")
	}

	return accounts.Account{
		ID:        acc.ID,
		Currency:  acc.Currency,
		Status:    acc.Status,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
	}, nil
}
