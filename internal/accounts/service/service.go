package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/accounts"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

type Service struct {
	store accounts.Store
}

var _ accounts.Service = (*Service)(nil)

func NewService(store accounts.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (accounts.Account, error) {
	return s.store.GetByID(ctx, id)
}

func (s Service) GetAll(ctx context.Context) ([]accounts.Account, error) {
	return s.store.GetAll(ctx)
}

func (s Service) Create(ctx context.Context, name string, currency money.Currency) (accounts.Account, error) {
	if name == "" {
		name = string(currency) + " account"
	}

	now := time.Now()
	acc := accounts.Account{
		ID:        uuid.New(),
		Name:      name,
		Currency:  currency,
		Status:    accounts.AccountStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.Create(ctx, acc); err != nil {
		return accounts.Account{}, errors.New("couldn't save new account")
	}
	return acc, nil
}

func (s Service) Close(ctx context.Context, id uuid.UUID) error {
	acc, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}

	acc.Status = accounts.AccountStatusClosed
	acc.UpdatedAt = time.Now()

	if err := s.store.Update(ctx, acc); err != nil {
		return errors.Wrap(err, "couldn't update account")
	}
	return nil
}
