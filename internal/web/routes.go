package web

import (
	"net/http"

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
		// 'GET /' redirects to the current month page
		{methods: "GET", path: "/", handler: s.indexHandler},

		// API
		{methods: "GET", path: "/api/months", handler: s.GetMonth},
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
	}

	for _, r := range routes {
		router.Methods(r.methods).Path(r.path).HandlerFunc(r.handler)
	}
}
