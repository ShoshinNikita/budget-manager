package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

const (
	overviewTemplatePath = "./templates/overview.html"
	yearTemplatePath     = "./templates/year.html"
	monthTemplatePath    = "./templates/month.html"
	//
	searchSpendsTemplatePath = "./templates/search_spends.html"
	//
	errorPageTemplatePath = "./templates/error_page.html"
)

const (
	executeErrorMessage     = "Can't execute template"
	invalidURLMessagePrefix = "Invalid URL: "
	dbErrorMessagePrefix    = "DB error: "
)

// GET /overview
//
func (s Server) overviewPage(w http.ResponseWriter, r *http.Request) {
	if err := s.tplStore.Execute(r.Context(), overviewTemplatePath, w, nil); err != nil {
		s.processErrorWithPage(r.Context(), w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}

// GET /overview/{year:[0-9]+}
//
func (s Server) yearPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		s.processErrorWithPage(
			r.Context(), w, invalidURLMessagePrefix+"invalid year", http.StatusBadRequest, nil,
		)
		return
	}

	months, err := s.db.GetMonths(r.Context(), year)
	// Render the page even theare no months for passed year
	if err != nil && err != db.ErrYearNotExist {
		msg, code, err := s.parseDBError(err)
		s.processErrorWithPage(r.Context(), w, dbErrorMessagePrefix+msg, code, err)
		return
	}

	// Display all months. Months without data in DB have zero id

	allMonths := make([]*db.Month, 12)
	for month := time.January; month <= time.December; month++ {
		allMonths[month-1] = &db.Month{
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

	resp := struct {
		Year         int
		Months       []*db.Month
		AnnualIncome money.Money
		AnnualSpend  money.Money
		Result       money.Money
	}{
		Year:         year,
		Months:       allMonths,
		AnnualIncome: annualIncome,
		AnnualSpend:  annualSpend,
		Result:       result,
	}
	if err := s.tplStore.Execute(r.Context(), yearTemplatePath, w, resp); err != nil {
		s.processErrorWithPage(r.Context(), w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}

// GET /overview/{year:[0-9]+}/{month:[0-9]+}
//
func (s Server) monthPage(w http.ResponseWriter, r *http.Request) {
	year, ok := getYear(r)
	if !ok {
		s.processErrorWithPage(
			r.Context(), w, invalidURLMessagePrefix+"invalid year", http.StatusBadRequest, nil,
		)
		return
	}
	monthNumber, ok := getMonth(r)
	if !ok {
		s.processErrorWithPage(r.Context(), w, invalidURLMessagePrefix+"invalid month", http.StatusBadRequest, nil)
		return
	}

	monthID, err := s.db.GetMonthID(r.Context(), year, int(monthNumber))
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processErrorWithPage(r.Context(), w, dbErrorMessagePrefix+msg, code, err)
		return
	}

	// Process
	month, err := s.db.GetMonth(r.Context(), monthID)
	if err != nil {
		s.processErrorWithPage(
			r.Context(), w, dbErrorMessagePrefix+"couldn't get Month info", http.StatusInternalServerError, err,
		)
		return
	}

	spendTypes, err := s.db.GetSpendTypes(r.Context())
	if err != nil {
		s.processErrorWithPage(r.Context(), w, dbErrorMessagePrefix+"couldn't get list of Spend Types",
			http.StatusInternalServerError, err)
		return
	}

	resp := struct {
		*db.Month

		SpendTypes    []*db.SpendType
		ToShortMonth  func(time.Month) string
		SumSpendCosts func([]*db.Spend) money.Money
	}{
		Month:         month,
		SpendTypes:    spendTypes,
		ToShortMonth:  toShortMonth,
		SumSpendCosts: sumSpendCosts,
	}
	if err := s.tplStore.Execute(r.Context(), monthTemplatePath, w, resp); err != nil {
		s.processErrorWithPage(r.Context(), w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}

// GET /search/spends
//
// Query Params:
//   - title - spend title
//   - notes - spend notes
//   - min_cost - minimal const
//   - max_cost - maximal cost
//   - after - date in format 'yyyy-mm-dd'
//   - before - date in format 'yyyy-mm-dd'
//   - type_id - Spend Type id to search (can be passed multiple times: ?type_id=56&type_id=58)
//
// nolint:funlen
func (s Server) searchSpendsPage(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Parse the query

	// Parse Title and Notes
	title := r.FormValue("title")
	notes := r.FormValue("notes")

	// Parse Min and Max Costs
	minCost := func() money.Money {
		minCostParam := r.FormValue("min_cost")
		if minCostParam == "" {
			return 0
		}

		minCost, err := strconv.ParseFloat(minCostParam, 64)
		if err != nil {
			// Just log this error
			log.WithError(err).WithField("min_cost", minCostParam).Warn("couldn't parse 'min_cost' param")
			return 0
		}
		return money.FromFloat(minCost)
	}()
	maxCost := func() money.Money {
		maxCostParam := r.FormValue("max_cost")
		if maxCostParam == "" {
			return 0
		}

		maxCost, err := strconv.ParseFloat(maxCostParam, 64)
		if err != nil {
			// Just log this error
			log.WithError(err).WithField("max_cost", maxCostParam).Warn("couldn't parse 'max_cost' param")
			return 0
		}
		return money.FromFloat(maxCost)
	}()

	// Parse After and Before
	const timeLayout = "2006-01-02"
	after := func() time.Time {
		after := r.FormValue("after")
		if after == "" {
			return time.Time{}
		}

		t, err := time.Parse(timeLayout, after)
		if err != nil {
			// Just log this error
			log.WithError(err).WithField("after", after).Warn("couldn't parse 'after' param")
			t = time.Time{}
		}
		return t
	}()
	before := func() time.Time {
		before := r.FormValue("before")
		if before == "" {
			return time.Time{}
		}

		t, err := time.Parse(timeLayout, before)
		if err != nil {
			// Just log this error
			log.WithError(err).WithField("before", before).Warn("couldn't parse 'before' param")
			t = time.Time{}
		}
		return t
	}()

	// Parse Spend Type ids
	typeIDs := func() []uint {
		ids := r.Form["type_id"]
		typeIDs := make([]uint, 0, len(ids))
		for i := range ids {
			id, err := strconv.ParseUint(ids[i], 10, 0)
			if err != nil {
				// Just log the error
				log.WithError(err).WithField("type_id", ids[i]).Warn("couldn't convert Spend Type id")
				continue
			}
			typeIDs = append(typeIDs, uint(id))
		}
		return typeIDs
	}()

	// Process
	args := db.SearchSpendsArgs{
		Title:   strings.ToLower(title),
		Notes:   strings.ToLower(notes),
		After:   after,
		Before:  before,
		MinCost: minCost,
		MaxCost: maxCost,
		TypeIDs: typeIDs,
		// TODO
		TitleExactly: false,
		NotesExactly: false,
	}
	spends, err := s.db.SearchSpends(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processErrorWithPage(r.Context(), w, msg, code, err)
		return
	}

	spendTypes, err := s.db.GetSpendTypes(r.Context())
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processErrorWithPage(r.Context(), w, msg, code, err)
		return
	}

	// Execute the template
	resp := struct {
		Spends     []*db.Spend
		SpendTypes []*db.SpendType
		TotalCost  money.Money
	}{
		Spends:     spends,
		SpendTypes: spendTypes,
		TotalCost:  sumSpendCosts(spends),
	}
	if err := s.tplStore.Execute(r.Context(), searchSpendsTemplatePath, w, resp); err != nil {
		s.processErrorWithPage(r.Context(), w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Helpers
// -------------------------------------------------

func toShortMonth(m time.Month) string {
	month := m.String()
	// Don't trim June and July
	if len(month) > 4 {
		month = m.String()[:3]
	}
	return month
}

func sumSpendCosts(spends []*db.Spend) money.Money {
	var m money.Money
	for i := range spends {
		m = m.Sub(spends[i].Cost)
	}
	return m
}

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
