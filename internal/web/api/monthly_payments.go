package api

import "github.com/sirupsen/logrus"

type MonthlyPaymentsHandlers struct {
	db  MonthlyPaymentsDB
	log logrus.FieldLogger
}

type MonthlyPaymentsDB interface {
}
