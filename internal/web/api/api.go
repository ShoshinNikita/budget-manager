package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

type API struct {
	service app.Service
	log     logger.Logger

	version   string
	gitHash   string
	startTime time.Time
}

func New(service app.Service, log logger.Logger, version, gitHash string) *API {
	return &API{
		service: service,
		log:     log,
		//
		version:   version,
		gitHash:   gitHash,
		startTime: time.Now(),
	}
}

func (api API) RegisterHandlers(mux *http.ServeMux) {
	writeError := func(w http.ResponseWriter, statusCode int) {
		w.WriteHeader(statusCode)
		fmt.Fprint(w, http.StatusText(statusCode))
	}

	for pattern, handler := range map[string]http.Handler{
		"/api/info/get": NewHandler(api.log, api.getInfo),
		//
		"/api/accounts/get":    NewHandler(api.log, api.getAccounts),
		"/api/accounts/create": NewHandler(api.log, api.createAccount),
		"/api/accounts/close":  NewHandler(api.log, api.closeAccount),
		//
		"/api/transactions/get":             NewHandler(api.log, api.getTransactions),
		"/api/transactions/create":          NewHandler(api.log, api.createTransaction),
		"/api/transactions/create/transfer": NewHandler(api.log, api.createTransferTransaction),
		"/api/transactions/delete":          NewHandler(api.log, api.deleteTransactions),
		//
		"/api/categories/get":    NewHandler(api.log, api.getCategories),
		"/api/categories/create": NewHandler(api.log, api.createCategory),
		"/api/categories/delete": NewHandler(api.log, api.deleteCategory),
	} {
		pattern := pattern
		handler := handler

		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != pattern {
				writeError(w, http.StatusNotFound)
				return
			}
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
