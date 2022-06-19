package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

func (s Service) GetTransactions(ctx context.Context, args app.GetTransactionsArgs) ([]app.Transaction, error) {
	return s.transactionStore.Get(ctx, args)
}

func (s Service) CalculateAccountBalances(
	ctx context.Context, accountIDs []uuid.UUID,
) (map[uuid.UUID]money.Money, error) {

	// TODO: use filter to get transactions only for required accounts?
	allTransaction, err := s.GetTransactions(ctx, app.GetTransactionsArgs{
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
		case app.TransactionTypeAdd:
			balance = balance.Add(t.Amount)
		case app.TransactionTypeWithdraw:
			balance = balance.Sub(t.Amount)
		}
		res[t.AccountID] = balance
	}
	return res, nil
}

func (s Service) CreateTransaction(
	ctx context.Context, args app.CreateTransactionArgs,
) (app.Transaction, error) {

	_, err := s.GetAccountByID(ctx, args.AccountID)
	if err != nil {
		return app.Transaction{}, errors.Wrap(err, "couldn't get account by id")
	}

	t := app.Transaction{
		ID:          uuid.New(),
		AccountID:   args.AccountID,
		Type:        args.Type,
		Name:        args.Name,
		Description: args.Description,
		Amount:      args.Amount,
		CategoryID:  args.CategoryID,
		CreatedAt:   time.Now(),
	}
	if err := s.transactionStore.Create(ctx, t); err != nil {
		return app.Transaction{}, errors.Wrap(err, "couldn't save new transaction")
	}
	return t, nil
}

func (s Service) CreateTransferTransactions(
	ctx context.Context, args app.CreateTransferTransactionsArgs,
) ([2]app.Transaction, error) {

	fromAccount, err := s.GetAccountByID(ctx, args.FromAccountID)
	if err != nil {
		return [2]app.Transaction{}, errors.Wrap(err, "couldn't get 'from' account")
	}
	toAccount, err := s.GetAccountByID(ctx, args.ToAccountID)
	if err != nil {
		return [2]app.Transaction{}, errors.Wrap(err, "couldn't get 'to account")
	}

	extra := &app.TransferTransactionExtra{
		TransferID: uuid.New(),
	}
	now := time.Now()

	name := fmt.Sprintf("Transfer %s", extra.TransferID)
	// TODO: format amount
	desc := fmt.Sprintf(
		"Transfer of %s %s to %s %s",
		args.FromAmount, fromAccount.Currency, args.ToAmount, toAccount.Currency,
	)

	transferTransactions := [2]app.Transaction{
		{
			ID:          uuid.New(),
			AccountID:   args.FromAccountID,
			Type:        app.TransactionTypeWithdraw,
			Flags:       app.TransactionFlagTransfer,
			Name:        name,
			Description: desc,
			Amount:      args.FromAmount,
			Extra:       extra,
			CreatedAt:   now,
		},
		{
			ID:          uuid.New(),
			AccountID:   args.ToAccountID,
			Type:        app.TransactionTypeAdd,
			Flags:       app.TransactionFlagTransfer,
			Name:        name,
			Description: desc,
			Amount:      args.ToAmount,
			Extra:       extra,
			CreatedAt:   now,
		},
	}
	if err := s.transactionStore.Create(ctx, transferTransactions[:]...); err != nil {
		return [2]app.Transaction{}, errors.Wrap(err, "couldn't save new transactions")
	}
	return transferTransactions, nil
}

func (s Service) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	transaction, err := s.transactionStore.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if transaction.Flags.IsTransferTransaction() {
		return errors.New("transfer transactions can't be deleted")
	}

	now := time.Now()
	transaction.DeletedAt = &now

	if err := s.transactionStore.Update(ctx, transaction); err != nil {
		return errors.Wrap(err, "couldn't update transaction for deletion")
	}
	return nil
}
