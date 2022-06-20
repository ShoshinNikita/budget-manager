package web

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

func (s Server) addRoutes(mux *http.ServeMux) {
	writeError := func(w http.ResponseWriter, statusCode int) {
		w.WriteHeader(statusCode)
		fmt.Fprint(w, http.StatusText(statusCode))
	}

	// Register API handlers
	for pattern, routes := range map[string]map[string]http.HandlerFunc{} {
		pattern := pattern
		routes := routes
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != pattern {
				writeError(w, http.StatusNotFound)
				return
			}

			handler, ok := routes[r.Method]
			if !ok {
				writeError(w, http.StatusMethodNotAllowed)
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
