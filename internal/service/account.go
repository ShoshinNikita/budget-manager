package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

func (s Service) GetAccountByID(ctx context.Context, id uuid.UUID) (app.AccountWithBalance, error) {
	acc, err := s.accountStore.GetByID(ctx, id)
	if err != nil {
		return app.AccountWithBalance{}, err
	}

	balances, err := s.CalculateAccountBalances(ctx, []uuid.UUID{acc.ID})
	if err != nil {
		return app.AccountWithBalance{}, errors.Wrap(err, "couldn't calculate account balance")
	}
	balance, ok := balances[acc.ID]
	if !ok {
		return app.AccountWithBalance{}, errors.Errorf("balance for account %q was not calculated", acc.ID)
	}

	return app.AccountWithBalance{
		Account: acc,
		Balance: balance,
	}, nil
}

func (s Service) GetAccounts(ctx context.Context) ([]app.AccountWithBalance, error) {
	accounts, err := s.accountStore.Get(ctx)
	if err != nil {
		return nil, err
	}

	accountIDs := make([]uuid.UUID, 0, len(accounts))
	for _, acc := range accounts {
		accountIDs = append(accountIDs, acc.ID)
	}
	balances, err := s.CalculateAccountBalances(ctx, accountIDs)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't calculate account balances")
	}

	res := make([]app.AccountWithBalance, 0, len(accounts))
	for _, acc := range accounts {
		balance, ok := balances[acc.ID]
		if !ok {
			return nil, errors.Errorf("got no balance for account %q", acc.ID)
		}
		res = append(res, app.AccountWithBalance{
			Account: acc,
			Balance: balance,
		})
	}
	return res, nil
}

func (s Service) CreateAccount(ctx context.Context, name string, currency money.Currency) (app.Account, error) {
	name = strings.TrimSpace(name)
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
