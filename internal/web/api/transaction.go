package api

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/validator"
)

type (
	getTransactionsReq struct {
		// TODO: add params
	}
	getTransactionsResp struct {
		Transactions []app.Transaction `json:"transactions"`
	}
)

func (api API) getTransactions(ctx context.Context, req *getTransactionsReq) (*getTransactionsResp, error) {
	args := app.GetTransactionsArgs{
		// TODO: use params from the request
	}
	transactions, err := api.service.GetTransactions(ctx, args)
	if err != nil {
		return nil, err
	}

	return &getTransactionsResp{
		Transactions: transactions,
	}, nil
}

type (
	createTransactionReq struct {
		AccountID   uuid.UUID                            `json:"accountID"`
		Type        validator.Valid[app.TransactionType] `json:"type"`
		Date        validator.Valid[app.Date]            `json:"date"`
		Name        string                               `json:"name"`
		Description string                               `json:"description"`
		Amount      money.Money                          `json:"amount"`
		CategoryID  uuid.UUID                            `json:"categoryID"`
	}
	createTransactionResp struct {
		NewTransaction app.Transaction `json:"newTransaction"`
	}
)

func (api API) createTransaction(ctx context.Context, req *createTransactionReq) (*createTransactionResp, error) {
	args := app.CreateTransactionArgs{
		AccountID:   req.AccountID,
		Type:        req.Type.Get(),
		Date:        req.Date.Get(),
		Name:        req.Name,
		Description: req.Description,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
	}
	newTransaction, err := api.service.CreateTransaction(ctx, args)
	if err != nil {
		return nil, err
	}
	return &createTransactionResp{
		NewTransaction: newTransaction,
	}, nil
}

type (
	createTransferTransactionReq struct {
		Date          validator.Valid[app.Date] `json:"date"`
		FromAccountID uuid.UUID                 `json:"fromAccountID"`
		FromAmount    money.Money               `json:"fromAmount"`
		ToAccountID   uuid.UUID                 `json:"toAccountID"`
		ToAmount      money.Money               `json:"toAmount"`
	}
	createTransferTransactionResp struct {
		NewTransactions []app.Transaction `json:"newTransactions"`
	}
)

func (api API) createTransferTransaction(
	ctx context.Context, req *createTransferTransactionReq,
) (*createTransferTransactionResp, error) {

	args := app.CreateTransferTransactionsArgs{
		Date:          req.Date.Get(),
		FromAccountID: req.FromAccountID,
		FromAmount:    req.FromAmount,
		ToAccountID:   req.ToAccountID,
		ToAmount:      req.ToAmount,
	}
	newTransferTransactions, err := api.service.CreateTransferTransactions(ctx, args)
	if err != nil {
		return nil, err
	}
	return &createTransferTransactionResp{
		NewTransactions: newTransferTransactions[:],
	}, nil
}

type deleteTransactionsReq struct {
	// TODO: support multiple ids?
	ID uuid.UUID `json:"id"`
}

func (api API) deleteTransactions(ctx context.Context, req *deleteTransactionsReq) (*emptyResp, error) {
	err := api.service.DeleteTransaction(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &emptyResp{}, nil
}
