package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

type route struct {
	methods string
	path    string
	handler http.Handler
}

func (s Server) addRoutes(router *mux.Router) {
	routes := []route{
		// Pages
		{methods: "GET", path: "/years", handler: http.HandlerFunc(notImplementedYet)},
		{methods: "GET", path: "/years/{year}", handler: http.HandlerFunc(notImplementedYet)},
		{methods: "GET", path: "/years/{year}/months", handler: http.HandlerFunc(notImplementedYet)},
		{methods: "GET", path: "/years/{year}/months/{month}", handler: http.HandlerFunc(notImplementedYet)},

		// API
		{methods: "GET", path: "/api/months", handler: http.HandlerFunc(s.GetMonth)},
		{methods: "GET", path: "/api/days", handler: http.HandlerFunc(s.GetDay)},
		// Income
		{methods: "POST", path: "/api/incomes", handler: http.HandlerFunc(s.AddIncome)},
		{methods: "PUT", path: "/api/incomes", handler: http.HandlerFunc(s.EditIncome)},
		{methods: "DELETE", path: "/api/incomes", handler: http.HandlerFunc(s.RemoveIncome)},
		// Monthly Payment
		{methods: "POST", path: "/api/monthly-payments", handler: http.HandlerFunc(s.AddMonthlyPayment)},
		{methods: "PUT", path: "/api/monthly-payments", handler: http.HandlerFunc(s.EditMonthlyPayment)},
		{methods: "DELETE", path: "/api/monthly-payments", handler: http.HandlerFunc(s.RemoveMonthlyPayment)},
		// Spend
		{methods: "POST", path: "/api/spends", handler: http.HandlerFunc(s.AddSpend)},
		{methods: "PUT", path: "/api/spends", handler: http.HandlerFunc(s.EditSpend)},
		{methods: "DELETE", path: "/api/spends", handler: http.HandlerFunc(s.RemoveSpend)},
		// Spend Type
		{methods: "GET", path: "/api/spend-types", handler: http.HandlerFunc(s.GetSpendTypes)},
		{methods: "POST", path: "/api/spend-types", handler: http.HandlerFunc(s.AddSpendType)},
		{methods: "PUT", path: "/api/spend-types", handler: http.HandlerFunc(s.EditSpendType)},
		{methods: "DELETE", path: "/api/spend-types", handler: http.HandlerFunc(s.RemoveSpendType)},
	}

	for _, r := range routes {
		router.Methods(r.methods).Path(r.path).Handler(r.handler)
	}
}
