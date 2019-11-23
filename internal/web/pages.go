package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /years
//
func (s Server) yearsPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(yearsTemplatePath, w, nil); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /years/{year:[0-9]+}
//
func (s Server) yearPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid year was passed", http.StatusBadRequest, nil)
		return
	}

	s.log.Debug(year)

	if err := s.tplStore.Execute(yearTemplatePath, w, nil); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /years/{year:[0-9]+}/months
//
func (s Server) monthsPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid year was passed", http.StatusBadRequest, nil)
		return
	}

	s.log.Debug(year)

	if err := s.tplStore.Execute(monthsTemplatePath, w, nil); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /years/{year:[0-9]+}/months/{month:[0-9]+}
//
func (s Server) monthPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid year was passed", http.StatusBadRequest, nil)
		return
	}
	month, ok := getMonth(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid month was passed", http.StatusBadRequest, nil)
		return
	}

	s.log.Debug(year, month)

	if err := s.tplStore.Execute(monthTemplatePath, w, nil); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Helpers
// -------------------------------------------------

const yearKey = "year"

func getYear(r *http.Request) (year int, ok bool) {
	s, ok := mux.Vars(r)[yearKey]
	if !ok {
		return 0, false
	}

	year, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return year, true
}

const monthKey = "month"

func getMonth(r *http.Request) (month time.Month, ok bool) {
	monthStr, ok := mux.Vars(r)[monthKey]
	if !ok {
		return 0, false
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		return 0, false
	}

	month = time.Month(monthInt)
	if !(time.January <= month && month <= time.December) {
		return 0, false
	}

	return month, true
}
