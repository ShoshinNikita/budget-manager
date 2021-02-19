package db

import (
	"errors"
)

// Common errors
var (
	ErrMonthNotExist          = errors.New("such Month doesn't exist")
	ErrDayNotExist            = errors.New("such Day doesn't exist")
	ErrIncomeNotExist         = errors.New("such Income doesn't exist")
	ErrMonthlyPaymentNotExist = errors.New("such Monthly Payment doesn't exist")
	ErrSpendNotExist          = errors.New("such Spend doesn't exist")
	ErrSpendTypeNotExist      = errors.New("such Spend Type doesn't exist")
)
