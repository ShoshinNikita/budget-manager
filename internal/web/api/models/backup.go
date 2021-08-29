package models

import "github.com/ShoshinNikita/budget-manager/internal/db"

type BackupResp struct {
	BaseResponse

	Backup Backup `json:"backup"`
}

type Backup struct {
	Incomes         []db.Income         `json:"incomes"`
	MonthlyPayments []db.MonthlyPayment `json:"monthly_payments"`
	Spends          []db.Spend          `json:"spends"`
	SpendTypes      []db.SpendType      `json:"spend_types"`
}
