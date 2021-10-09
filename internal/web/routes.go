package web

import (
	"errors"
	"net/http"
	"net/http/pprof"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

//nolint:funlen
func (s Server) addRoutes(mux *http.ServeMux) {
	var (
		errUnknownPath      = errors.New("unknown path")
		errMethodNowAllowed = errors.New("method not allowed")
	)
	writeUnknownPathError := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := reqid.FromContextToLogger(ctx, s.log)

		utils.EncodeError(ctx, w, log, errUnknownPath, http.StatusNotFound)
	}
	writeMethodNowAllowedError := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := reqid.FromContextToLogger(ctx, s.log)

		utils.EncodeError(ctx, w, log, errMethodNowAllowed, http.StatusMethodNotAllowed)
	}

	pageHandlers := pages.NewHandlers(s.db, s.log, s.config.UseEmbed, s.version, s.gitHash)

	// Register the main handler. It serves pages and handles all requests with an unrecognized pattern
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var handler http.HandlerFunc
		switch r.URL.Path {
		case "/":
			handler = pageHandlers.IndexPage
		case "/months":
			handler = pageHandlers.MonthsPage
		case "/months/month":
			handler = pageHandlers.MonthPage
		case "/search/spends":
			handler = pageHandlers.SearchSpendsPage
		default:
			writeUnknownPathError(w, r)
			return
		}

		// Only GET requests are allowed
		if r.Method != http.MethodGet {
			writeMethodNowAllowedError(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	})

	apiHandlers := api.NewHandlers(s.db, s.log)

	// Register API handlers
	for pattern, routes := range map[string]map[string]http.HandlerFunc{
		"/api/months/date": {
			http.MethodGet: apiHandlers.GetMonthByDate,
		},
		"/api/incomes": {
			http.MethodPost:   apiHandlers.AddIncome,
			http.MethodPut:    apiHandlers.EditIncome,
			http.MethodDelete: apiHandlers.RemoveIncome,
		},
		"/api/monthly-payments": {
			http.MethodPost:   apiHandlers.AddMonthlyPayment,
			http.MethodPut:    apiHandlers.EditMonthlyPayment,
			http.MethodDelete: apiHandlers.RemoveMonthlyPayment,
		},
		"/api/spends": {
			http.MethodPost:   apiHandlers.AddSpend,
			http.MethodPut:    apiHandlers.EditSpend,
			http.MethodDelete: apiHandlers.RemoveSpend,
		},
		"/api/spend-types": {
			http.MethodGet:    apiHandlers.GetSpendTypes,
			http.MethodPost:   apiHandlers.AddSpendType,
			http.MethodPut:    apiHandlers.EditSpendType,
			http.MethodDelete: apiHandlers.RemoveSpendType,
		},
		"/api/search/spends": {
			http.MethodGet: apiHandlers.SearchSpends,
		},
	} {
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
