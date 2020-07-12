package web

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"

	"github.com/ShoshinNikita/budget-manager/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages/templates"
)

type route struct {
	methods string
	path    string
	handler http.HandlerFunc
}

func (s Server) addRoutes(router *mux.Router) {
	executorLog := s.log.WithField("component", "template_executor")
	var tplExecutor pages.TemplateExecutor = templates.NewTemplateDiskExecutor(executorLog)
	if s.config.CacheTemplates {
		tplExecutor = templates.NewTemplateCacheExecutor(executorLog)
	}
	pageHandlers := pages.NewHandlers(s.db, tplExecutor, s.log)

	apiHandlers := api.NewHandlers(s.db, s.log)

	routes := []route{
		// Pages
		{methods: "GET", path: "/overview", handler: pageHandlers.OverviewPage},
		{methods: "GET", path: "/overview/{year:[0-9]+}", handler: pageHandlers.YearPage},
		{methods: "GET", path: "/overview/{year:[0-9]+}/{month:[0-9]+}", handler: pageHandlers.MonthPage},
		{methods: "GET", path: "/search/spends", handler: pageHandlers.SearchSpendsPage},
		// 'GET /' redirects to the current month page
		{methods: "GET", path: "/", handler: s.indexHandler},

		// API
		{methods: "GET", path: "/api/months/id", handler: apiHandlers.GetMonthByID},
		{methods: "GET", path: "/api/months/date", handler: apiHandlers.GetMonthByDate},
		{methods: "GET", path: "/api/days/id", handler: apiHandlers.GetDayByID},
		{methods: "GET", path: "/api/days/date", handler: apiHandlers.GetDayByDate},
		// Income
		{methods: "POST", path: "/api/incomes", handler: apiHandlers.AddIncome},
		{methods: "PUT", path: "/api/incomes", handler: apiHandlers.EditIncome},
		{methods: "DELETE", path: "/api/incomes", handler: apiHandlers.RemoveIncome},
		// Monthly Payment
		{methods: "POST", path: "/api/monthly-payments", handler: apiHandlers.AddMonthlyPayment},
		{methods: "PUT", path: "/api/monthly-payments", handler: apiHandlers.EditMonthlyPayment},
		{methods: "DELETE", path: "/api/monthly-payments", handler: apiHandlers.RemoveMonthlyPayment},
		// Spend
		{methods: "POST", path: "/api/spends", handler: apiHandlers.AddSpend},
		{methods: "PUT", path: "/api/spends", handler: apiHandlers.EditSpend},
		{methods: "DELETE", path: "/api/spends", handler: apiHandlers.RemoveSpend},
		// Spend Type
		{methods: "GET", path: "/api/spend-types", handler: apiHandlers.GetSpendTypes},
		{methods: "POST", path: "/api/spend-types", handler: apiHandlers.AddSpendType},
		{methods: "PUT", path: "/api/spend-types", handler: apiHandlers.EditSpendType},
		{methods: "DELETE", path: "/api/spend-types", handler: apiHandlers.RemoveSpendType},
		// Other
		{methods: "GET", path: "/api/search/spends", handler: apiHandlers.SearchSpends},
	}
	for _, r := range routes {
		router.Methods(r.methods).Path(r.path).HandlerFunc(r.handler)
	}
}

func (Server) addPprofRoutes(router *mux.Router) {
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
}
