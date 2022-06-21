package api

import (
	"context"

	"github.com/google/uuid"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/validator"
)

type getAccountsResp struct {
	Accounts []app.Account `json:"accounts"`
}

func (api API) getAccounts(ctx context.Context, req *emptyReq) (*getAccountsResp, error) {
	accounts, err := api.service.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return &getAccountsResp{
		Accounts: accounts,
	}, nil
}

type (
	createAccountsReq struct {
		Name     string                          `json:"name"`
		Currency validator.Valid[money.Currency] `json:"currency"`
	}
	createAccountsResp struct {
		NewAccount app.Account `json:"newAccount"`
	}
)

func (api API) createAccount(ctx context.Context, req *createAccountsReq) (*createAccountsResp, error) {
	newAccount, err := api.service.CreateAccount(ctx, req.Name, req.Currency.Get())
	if err != nil {
		return nil, err
	}
	return &createAccountsResp{
		NewAccount: newAccount,
	}, nil
}

type closeAccountsReq struct {
	ID uuid.UUID `json:"id"`
}

func (api API) closeAccount(ctx context.Context, req *closeAccountsReq) (*emptyResp, error) {
	err := api.service.CloseAccount(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &emptyResp{}, nil
}
