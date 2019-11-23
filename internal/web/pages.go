package web

import (
	"net/http"
)

const (
	indexTemplatePath  = "./templates/index.html"
	yearsTemplatePath  = "./templates/years.html"
	yearTemplatePath   = "./templates/year.html"
	monthsTemplatePath = "./templates/months.html"
	monthTemplatePath  = "./templates/month.html"
)

// GET /
//
func (s Server) indexPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(indexTemplatePath, w, nil); err != nil {
		http.Error(w, "can't load template", http.StatusInternalServerError)
	}
}

// GET /years
//
func (s Server) yearsPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(yearsTemplatePath, w, nil); err != nil {
		http.Error(w, "can't load template", http.StatusInternalServerError)
	}
}

// GET /years/{year:[0-9]+}
//
func (s Server) yearPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(yearTemplatePath, w, nil); err != nil {
		http.Error(w, "can't load template", http.StatusInternalServerError)
	}
}

// GET /years/{year:[0-9]+}/months
//
func (s Server) monthsPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(monthsTemplatePath, w, nil); err != nil {
		http.Error(w, "can't load template", http.StatusInternalServerError)
	}
}

// GET /years/{year:[0-9]+}/months/{month:[0-9]+}
//
func (s Server) monthPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(monthTemplatePath, w, nil); err != nil {
		http.Error(w, "can't load template", http.StatusInternalServerError)
	}
}
