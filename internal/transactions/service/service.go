package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
)

type Service struct {
	store transactions.Store
}

var _ transactions.Service = (*Service)(nil)

func NewService(store transactions.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) Get(ctx context.Context, args transactions.GetTransactionsArgs) ([]transactions.Transaction, error) {
	return s.store.Get(ctx, args)
}

func (s Service) CalculateAccountBalances(
	ctx context.Context, accountIDs []uuid.UUID,
) (map[uuid.UUID]money.Money, error) {

	// TODO: use filter to get transactions only for required accounts?
	allTransaction, err := s.Get(ctx, transactions.GetTransactionsArgs{})
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

	t := transactions.Transaction{
		ID:         uuid.New(),
		AccountID:  args.AccountID,
		Type:       args.Type,
		Amount:     args.Amount,
		CategoryID: args.CategoryID,
		CreatedAt:  time.Now(),
	}
	if err := s.store.Create(ctx, t); err != nil {
		return transactions.Transaction{}, errors.Wrap(err, "couldn't save new transaction")
	}
	return t, nil
}

func (s Service) CreateTransferTransactions(
	ctx context.Context, args transactions.CreateTransferTransactionsArgs,
) ([2]transactions.Transaction, error) {

	extra := &transactions.TransferTransactionExtra{
		TransferID: uuid.New(),
	}
	now := time.Now()

	transferTransactions := [2]transactions.Transaction{
		{
			ID:        uuid.New(),
			AccountID: args.FromAccountID,
			Type:      transactions.TransactionTypeWithdraw,
			Flags:     transactions.TransactionFlagTransfer,
			Amount:    args.FromAmount,
			Extra:     extra,
			CreatedAt: now,
		},
		{
			ID:        uuid.New(),
			AccountID: args.ToAccountID,
			Type:      transactions.TransactionTypeAdd,
			Flags:     transactions.TransactionFlagTransfer,
			Amount:    args.ToAmount,
			Extra:     extra,
			CreatedAt: now,
		},
	}
	if err := s.store.Create(ctx, transferTransactions[:]...); err != nil {
		return [2]transactions.Transaction{}, errors.Wrap(err, "couldn't save new transactions")
	}
	return transferTransactions, nil
}
