package db

import (
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// Common errors
var (
	ErrYearNotExist = errors.New("there're no records for passed year",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrMonthNotExist = errors.New("Month with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrDayNotExist = errors.New("Day with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrIncomeNotExist = errors.New("Income with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrMonthlyPaymentNotExist = errors.New("Monthly Payment with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrSpendNotExist = errors.New("Spend with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
	ErrSpendTypeNotExist = errors.New("Spend Type with passed id doesn't exist",
		errors.WithOriginalError(), errors.WithType(errors.UserError),
	)
)
