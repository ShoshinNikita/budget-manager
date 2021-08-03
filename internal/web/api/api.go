package api

import "github.com/sirupsen/logrus"

type Handlers struct {
	MonthsHandlers
	IncomesHandlers
	MonthlyPaymentsHandlers
	SpendsHandlers
	SpendTypesHandlers
	SearchHandlers
}

type DB interface {
	MonthsDB
	IncomesDB
	MonthlyPaymentsDB
	SpendsDB
	SpendTypesDB
	SearchDB
}

func NewHandlers(db DB, log logrus.FieldLogger) *Handlers {
	return &Handlers{
		MonthsHandlers:          MonthsHandlers{db: db, log: log},
		IncomesHandlers:         IncomesHandlers{db: db, log: log},
		MonthlyPaymentsHandlers: MonthlyPaymentsHandlers{db: db, log: log},
		SpendsHandlers:          SpendsHandlers{db: db, log: log},
		SpendTypesHandlers:      SpendTypesHandlers{db: db, log: log},
		SearchHandlers:          SearchHandlers{db: db, log: log},
	}
}
