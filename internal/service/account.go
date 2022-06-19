package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

func (s Service) GetAccountByID(ctx context.Context, id uuid.UUID) (app.Account, error) {
	return s.accountStore.GetByID(ctx, id)
}

func (s Service) GetAccounts(ctx context.Context) ([]app.Account, error) {
	return s.accountStore.Get(ctx)
}

func (s Service) CreateAccount(ctx context.Context, name string, currency money.Currency) (app.Account, error) {
	if name == "" {
		name = string(currency) + " account"
	}

	now := time.Now()
	acc := app.Account{
		ID:        uuid.New(),
		Name:      name,
		Currency:  currency,
		Status:    app.AccountStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.accountStore.Create(ctx, acc); err != nil {
		return app.Account{}, errors.New("couldn't save new account")
	}
	return acc, nil
}

func (s Service) CloseAccount(ctx context.Context, id uuid.UUID) error {
	acc, err := s.accountStore.GetByID(ctx, id)
	if err != nil {
		return err
	}

	acc.Status = app.AccountStatusClosed
	acc.UpdatedAt = time.Now()

	if err := s.accountStore.Update(ctx, acc); err != nil {
		return errors.Wrap(err, "couldn't update account")
	}
	return nil
}
