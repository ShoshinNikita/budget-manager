package web

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

type route struct {
	methods string
	path    string
	handler http.HandlerFunc
}

func (s Server) addRoutes(router *mux.Router) {
	routes := []route{
		// Pages
		{methods: "GET", path: "/overview", handler: s.overviewPage},
		{methods: "GET", path: "/overview/{year:[0-9]+}", handler: s.yearPage},
		{methods: "GET", path: "/overview/{year:[0-9]+}/{month:[0-9]+}", handler: s.monthPage},
		{methods: "GET", path: "/search/spends", handler: s.searchSpendsPage},
		// 'GET /' redirects to the current month page
		{methods: "GET", path: "/", handler: s.indexHandler},

		// API
		{methods: "GET", path: "/api/months/id", handler: s.GetMonthByID},
		{methods: "GET", path: "/api/months/date", handler: s.GetMonthByDate},
		{methods: "GET", path: "/api/days", handler: s.GetDay},
		// Income
		{methods: "POST", path: "/api/incomes", handler: s.AddIncome},
		{methods: "PUT", path: "/api/incomes", handler: s.EditIncome},
		{methods: "DELETE", path: "/api/incomes", handler: s.RemoveIncome},
		// Monthly Payment
		{methods: "POST", path: "/api/monthly-payments", handler: s.AddMonthlyPayment},
		{methods: "PUT", path: "/api/monthly-payments", handler: s.EditMonthlyPayment},
		{methods: "DELETE", path: "/api/monthly-payments", handler: s.RemoveMonthlyPayment},
		// Spend
		{methods: "POST", path: "/api/spends", handler: s.AddSpend},
		{methods: "PUT", path: "/api/spends", handler: s.EditSpend},
		{methods: "DELETE", path: "/api/spends", handler: s.RemoveSpend},
		// Spend Type
		{methods: "GET", path: "/api/spend-types", handler: s.GetSpendTypes},
		{methods: "POST", path: "/api/spend-types", handler: s.AddSpendType},
		{methods: "PUT", path: "/api/spend-types", handler: s.EditSpendType},
		{methods: "DELETE", path: "/api/spend-types", handler: s.RemoveSpendType},
		// Other
		{methods: "GET", path: "/api/search/spends", handler: s.SearchSpends},
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
