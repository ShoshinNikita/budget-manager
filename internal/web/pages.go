package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/models"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

const (
	overviewTemplatePath = "./templates/overview.html"
	yearTemplatePath     = "./templates/year.html"
	monthTemplatePath    = "./templates/month.html"
)

// GET /overview
//
func (s Server) overviewPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(overviewTemplatePath, w, nil); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /overview/{year:[0-9]+}
//
func (s Server) yearPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid year was passed", http.StatusBadRequest, nil)
		return
	}

	months, err := s.db.GetMonths(context.Background(), year)
	// Render the page even theare no months for passed year
	if err != nil && err != db.ErrYearNotExist {
		s.processDBError(w, err)
		return
	}

	// Display all months. Months without data in DB have zero id

	allMonths := make([]*models.Month, 12)
	for month := time.January; month <= time.December; month++ {
		allMonths[month-1] = &models.Month{
			ID:    0,
			Year:  year,
			Month: month,
		}
	}
	for _, m := range months {
		allMonths[m.Month-1] = m
	}

	annualIncome := func() money.Money {
		var res money.Money
		for _, m := range allMonths {
			res = res.Add(m.TotalIncome)
		}
		return res
	}()

	annualSpend := func() money.Money {
		var res money.Money
		for _, m := range allMonths {
			// Use Add because 'TotalSpend' is negative
			res = res.Add(m.TotalSpend)
		}
		return res
	}()

	// Use Add because 'annualSpend' is negative
	result := annualIncome.Add(annualSpend)

	resp := map[string]interface{}{
		"Year":         year,
		"Months":       allMonths,
		"AnnualIncome": annualIncome,
		"AnnualSpend":  annualSpend,
		"Result":       result,
	}
	if err := s.tplStore.Execute(yearTemplatePath, w, resp); err != nil {
		// TODO: use special 500 page
		s.processError(w, "can't load template", http.StatusInternalServerError, err)
	}
}

// GET /overview/{year:[0-9]+}/{month:[0-9]+}
//
func (s Server) monthPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid year was passed", http.StatusBadRequest, nil)
		return
	}
	monthNumber, ok := getMonth(r)
	if !ok {
		// TODO: use special 404 page
		s.processError(w, "invalid month was passed", http.StatusBadRequest, nil)
		return
	}

	monthID, err := s.db.GetMonthID(context.Background(), year, int(monthNumber))
	if err != nil {
		s.processDBError(w, err)
		return
	}

	// Process
	month, err := s.db.GetMonth(context.Background(), monthID)
	if err != nil {
		s.processError(w, "can't select Month info", http.StatusInternalServerError, err)
		return
	}

	spendTypes, err := s.db.GetSpendTypes(context.Background())
	if err != nil {
		s.processError(w, "can't get Spend Types", http.StatusInternalServerError, err)
		return
	}

	resp := struct {
		*models.Month

		SpendTypes   []models.SpendType
		ToShortMonth func(time.Month) string
	}{
		Month:        month,
		SpendTypes:   spendTypes,
		ToShortMonth: toShortMonth,
	}
	if err := s.tplStore.Execute(monthTemplatePath, w, resp); err != nil {
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

func toShortMonth(m time.Month) string {
	month := m.String()
	// Don't trim June and July
	if len(month) > 4 {
		month = m.String()[:3]
	}
	return month
}
