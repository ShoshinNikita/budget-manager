// Package service contains all domain logic
package service

import "github.com/ShoshinNikita/budget-manager/v2/internal/app"

type Service struct {
	accountStore     app.AccountStore
	transactionStore app.TransactionStore
	categoryStore    app.CategoryStore
}

func NewService(
	accountStore app.AccountStore, transactionStore app.TransactionStore, categoryStore app.CategoryStore,
) *Service {

	return &Service{
		accountStore:     accountStore,
		transactionStore: transactionStore,
		categoryStore:    categoryStore,
	}
}
