package api

import (
	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

type API struct {
	service app.Service
	log     logger.Logger
}

func New(service app.Service, log logger.Logger) *API {
	return &API{
		service: service,
		log:     log,
	}
}
