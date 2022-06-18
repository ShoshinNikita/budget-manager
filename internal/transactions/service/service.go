package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/accounts"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
)

type Service struct {
	store           transactions.Store
	accountsService accounts.Service
}

var _ transactions.Service = (*Service)(nil)

func NewService(store transactions.Store, accountsService accounts.Service) *Service {
	return &Service{
		store:           store,
		accountsService: accountsService,
	}
}

func (s Service) Get(ctx context.Context, args transactions.GetTransactionsArgs) ([]transactions.Transaction, error) {
	return s.store.Get(ctx, args)
}

func (s Service) CalculateAccountBalances(
	ctx context.Context, accountIDs []uuid.UUID,
) (map[uuid.UUID]money.Money, error) {

	// TODO: use filter to get transactions only for required accounts?
	allTransaction, err := s.Get(ctx, transactions.GetTransactionsArgs{
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get transactions")
	}

	res := make(map[uuid.UUID]money.Money, len(accountIDs))
	for _, id := range accountIDs {
		res[id] = money.FromInt(0)
	}

	for _, t := range allTransaction {
		balance, ok := res[t.AccountID]
		if !ok {
			continue
		}

		switch t.Type {
		case transactions.TransactionTypeAdd:
			balance = balance.Add(t.Amount)
		case transactions.TransactionTypeWithdraw:
			balance = balance.Sub(t.Amount)
		}
		res[t.AccountID] = balance
	}
	return res, nil
}

func (s Service) CreateTransaction(
	ctx context.Context, args transactions.CreateTransactionArgs,
) (transactions.Transaction, error) {

	_, err := s.accountsService.GetByID(ctx, args.AccountID)
	if err != nil {
		return transactions.Transaction{}, errors.Wrap(err, "couldn't get account by id")
	}

	t := transactions.Transaction{
		ID:          uuid.New(),
		AccountID:   args.AccountID,
		Type:        args.Type,
		Name:        args.Name,
		Description: args.Description,
		Amount:      args.Amount,
		CategoryID:  args.CategoryID,
		CreatedAt:   time.Now(),
	}
	if err := s.store.Create(ctx, t); err != nil {
		return transactions.Transaction{}, errors.Wrap(err, "couldn't save new transaction")
	}
	return t, nil
}

func (s Service) CreateTransferTransactions(
	ctx context.Context, args transactions.CreateTransferTransactionsArgs,
) ([2]transactions.Transaction, error) {

	fromAccount, err := s.accountsService.GetByID(ctx, args.FromAccountID)
	if err != nil {
		return [2]transactions.Transaction{}, errors.Wrap(err, "couldn't get 'from' account")
	}
	toAccount, err := s.accountsService.GetByID(ctx, args.ToAccountID)
	if err != nil {
		return [2]transactions.Transaction{}, errors.Wrap(err, "couldn't get 'to account")
	}

	extra := &transactions.TransferTransactionExtra{
		TransferID: uuid.New(),
	}
	now := time.Now()

	name := fmt.Sprintf("Transfer %s", extra.TransferID)
	// TODO: format amount
	desc := fmt.Sprintf(
		"Transfer of %d %s to %d %s",
		args.FromAmount, fromAccount.Currency, args.ToAmount, toAccount.Currency,
	)

	transferTransactions := [2]transactions.Transaction{
		{
			ID:          uuid.New(),
			AccountID:   args.FromAccountID,
			Type:        transactions.TransactionTypeWithdraw,
			Flags:       transactions.TransactionFlagTransfer,
			Name:        name,
			Description: desc,
			Amount:      args.FromAmount,
			Extra:       extra,
			CreatedAt:   now,
		},
		{
			ID:          uuid.New(),
			AccountID:   args.ToAccountID,
			Type:        transactions.TransactionTypeAdd,
			Flags:       transactions.TransactionFlagTransfer,
			Name:        name,
			Description: desc,
			Amount:      args.ToAmount,
			Extra:       extra,
			CreatedAt:   now,
		},
	}
	if err := s.store.Create(ctx, transferTransactions[:]...); err != nil {
		return [2]transactions.Transaction{}, errors.Wrap(err, "couldn't save new transactions")
	}
	return transferTransactions, nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	transaction, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now()
	transaction.DeletedAt = &now

	if err := s.store.Update(ctx, transaction); err != nil {
		return errors.Wrap(err, "couldn't update transaction for deletion")
	}
	return nil
}
