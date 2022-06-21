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

	for pattern, handler := range map[string]http.HandlerFunc{
		"/api/info/get": api.getInfo,
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

func (api API) getInfo(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Version string `json:"version"`
		GitHash string `json:"gitHash"`
		Uptime  string `json:"uptime"`
	}{
		Version: api.version,
		GitHash: api.gitHash,
		Uptime:  time.Since(api.startTime).String(),
	}
	api.encodeResponse(r.Context(), w, http.StatusOK, resp)
}
