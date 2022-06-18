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

type AccountsStore struct {
	base *BaseStore[accounts.Account]
}

var _ accounts.Store = (*AccountsStore)(nil)

func NewAccountsStore(boltStore *bbolt.DB) (*AccountsStore, error) {
	store := &AccountsStore{
		base: NewBaseStore(
			boltStore, "accounts", marshalBoltAccount, unmarshalBoltAccount,
		),
	}

	if err := store.base.Init(); err != nil {
		return nil, errors.Wrap(err, "couldn't init store")
	}
	return store, nil
}

func (store AccountsStore) GetByID(ctx context.Context, id uuid.UUID) (accounts.Account, error) {
	return store.base.GetByID(id)
}

func (store AccountsStore) GetAll(ctx context.Context) ([]accounts.Account, error) {
	return store.base.GetAll(
		nil,
		func(accs []accounts.Account) {
			sort.Slice(accs, func(i, j int) bool {
				return accs[i].CreatedAt.Before(accs[j].CreatedAt)
			})
		},
	)
}

func (store AccountsStore) Create(ctx context.Context, acc accounts.Account) error {
	return store.base.Create(acc)
}

func (store AccountsStore) Update(ctx context.Context, acc accounts.Account) error {
	return store.base.Update(acc)
}

type boltAccount struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Currency  money.Currency         `json:"currency"`
	Status    accounts.AccountStatus `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func marshalBoltAccount(acc accounts.Account) []byte {
	data, err := json.Marshal(boltAccount{
		ID:        acc.ID,
		Name:      acc.Name,
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
		Name:      acc.Name,
		Currency:  acc.Currency,
		Status:    acc.Status,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
	}, nil
}
