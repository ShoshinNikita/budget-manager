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
		{methods: "GET", path: "/", handler: s.indexPage},
		{methods: "GET", path: "/years", handler: s.yearsPage},
		{methods: "GET", path: "/years/{year:[0-9]+}", handler: s.yearPage},
		{methods: "GET", path: "/years/{year:[0-9]+}/months", handler: s.monthsPage},
		{methods: "GET", path: "/years/{year:[0-9]+}/months/{month:[0-9]+}", handler: s.monthPage},

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
