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
		// Income
		{methods: "POST", path: "/api/incomes", handler: http.HandlerFunc(s.AddIncome)},
		{methods: "PUT", path: "/api/incomes/{id}", handler: http.HandlerFunc(s.EditIncome)},
		{methods: "DELETE", path: "/api/incomes/{id}", handler: http.HandlerFunc(s.DeleteIncome)},
		// Monthly Payment
		{methods: "POST", path: "/api/monthly-payments", handler: http.HandlerFunc(s.AddMonthlyPayment)},
		{methods: "PUT", path: "/api/monthly-payments/{id}", handler: http.HandlerFunc(s.EditMonthlyPayment)},
		{methods: "DELETE", path: "/api/monthly-payment/{id}", handler: http.HandlerFunc(s.DeleteMonthlyPayment)},
		// Spend
		{methods: "POST", path: "/api/spends", handler: http.HandlerFunc(s.AddSpend)},
		{methods: "PUT", path: "/api/spends/{id}", handler: http.HandlerFunc(s.EditSpend)},
		{methods: "DELETE", path: "/api/spends/{id}", handler: http.HandlerFunc(s.DeleteSpend)},
		// Spend Type
		{methods: "POST", path: "/api/spend-types", handler: http.HandlerFunc(s.AddSpend)},
		{methods: "PUT", path: "/api/spend-types/{id}", handler: http.HandlerFunc(s.EditSpend)},
		{methods: "DELETE", path: "/api/spend-types/{id}", handler: http.HandlerFunc(s.DeleteSpend)},
	}

	for _, r := range routes {
		router.Methods(r.methods).Path(r.path).Handler(r.handler)
	}
}
