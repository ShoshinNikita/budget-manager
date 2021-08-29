package api

import "github.com/ShoshinNikita/budget-manager/internal/logger"

type Handlers struct {
	MonthsHandlers
	IncomesHandlers
	MonthlyPaymentsHandlers
	SpendsHandlers
	SpendTypesHandlers
	SearchHandlers
	BackupHandlers
}

type DB interface {
	MonthsDB
	IncomesDB
	MonthlyPaymentsDB
	SpendsDB
	SpendTypesDB
	SearchDB
	BackupDB
}

func NewHandlers(db DB, log logger.Logger) *Handlers {
	return &Handlers{
		MonthsHandlers:          MonthsHandlers{db, log},
		IncomesHandlers:         IncomesHandlers{db, log},
		MonthlyPaymentsHandlers: MonthlyPaymentsHandlers{db, log},
		SpendsHandlers:          SpendsHandlers{db, log},
		SpendTypesHandlers:      SpendTypesHandlers{db, log},
		SearchHandlers:          SearchHandlers{db, log},
		BackupHandlers:          BackupHandlers{db, log},
	}
}
