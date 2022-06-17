package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/accounts"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
)

type Service struct {
	store               accounts.Store
	transactionsService transactions.Service
}

var _ accounts.Service = (*Service)(nil)

func NewService(store accounts.Store, transactionsService transactions.Service) *Service {
	return &Service{
		store:               store,
		transactionsService: transactionsService,
	}
}

func (s Service) GetAll(ctx context.Context) ([]accounts.AccountWithBalance, error) {
	allAccounts, err := s.store.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get all accounts from store")
	}

	allIDs := make([]uuid.UUID, 0, len(allAccounts))
	for _, acc := range allAccounts {
		allIDs = append(allIDs, acc.ID)
	}

	balances, err := s.transactionsService.CalculateAccountBalances(ctx, allIDs)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't calculate account balances")
	}

	res := make([]accounts.AccountWithBalance, 0, len(allAccounts))
	for _, acc := range allAccounts {
		balance, ok := balances[acc.ID]
		if !ok {
			return nil, errors.Errorf("transaction service return no balance for account %s", acc.ID)
		}

		res = append(res, accounts.AccountWithBalance{
			Account: acc,
			Balance: balance,
		})
	}
	return res, nil
}

func (s Service) Create(ctx context.Context, currency money.Currency) (accounts.Account, error) {
	now := time.Now()
	acc := accounts.Account{
		ID:        uuid.New(),
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
