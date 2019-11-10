package web

import "net/http"

// -------------------------------------------------
// Income
// -------------------------------------------------

func (s Server) AddIncome(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditIncome(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteIncome(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Monthly Payment
// -------------------------------------------------

func (s Server) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Spends
// -------------------------------------------------

func (s Server) AddSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Spend Types
// -------------------------------------------------

func (s Server) AddSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Other
// -------------------------------------------------

func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet"))
}
