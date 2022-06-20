package web

import (
	"errors"
	"net/http"
	"net/http/pprof"

	"github.com/ShoshinNikita/budget-manager/v2/internal/web/utils"
)

func (s Server) addRoutes(mux *http.ServeMux) {
	var (
		errUnknownPath      = errors.New("unknown path")
		errMethodNowAllowed = errors.New("method not allowed")
	)
	writeUnknownPathError := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		utils.EncodeError(ctx, w, s.log, errUnknownPath, http.StatusNotFound)
	}
	writeMethodNowAllowedError := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		utils.EncodeError(ctx, w, s.log, errMethodNowAllowed, http.StatusMethodNotAllowed)
	}

	// Register API handlers
	for pattern, routes := range map[string]map[string]http.HandlerFunc{} {
		pattern := pattern
		routes := routes
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != pattern {
				writeUnknownPathError(w, r)
				return
			}

			handler, ok := routes[r.Method]
			if !ok {
				writeMethodNowAllowedError(w, r)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

func (Server) addPprofRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
