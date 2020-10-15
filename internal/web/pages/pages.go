package pages

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/version"
)

const (
	executeErrorMessage     = "Can't execute template"
	invalidURLMessagePrefix = "Invalid URL: "
	dbErrorMessagePrefix    = "DB error: "
)

const (
	overviewTemplateName     = "overview.html"
	yearTemplateName         = "overview_year.html"
	monthTemplateName        = "overview_year_month.html"
	searchSpendsTemplateName = "search_spends.html"
	errorPageTemplateName    = "error_page.html"
)

type Handlers struct {
	db          DB
	tplExecutor *templateExecutor
	log         logrus.FieldLogger
}

type DB interface {
	GetMonth(ctx context.Context, id uint) (db.Month, error)
	GetMonthID(ctx context.Context, year, month int) (uint, error)
	GetMonths(ctx context.Context, year int) ([]db.Month, error)

	GetSpendTypes(ctx context.Context) ([]db.SpendType, error)

	SearchSpends(ctx context.Context, args db.SearchSpendsArgs) ([]db.Spend, error)
}

func NewHandlers(db DB, log logrus.FieldLogger, cacheTemplates bool) *Handlers {
	return &Handlers{
		db:          db,
		tplExecutor: newTemplateExecutor(log, cacheTemplates, commonTemplateFuncs()),
		log:         log,
	}
}

func commonTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"asStaticURL": func(rawURL string) (string, error) {
			url, err := url.Parse(rawURL)
			if err != nil {
				return "", err
			}

			query := url.Query()
			query.Add("hash", version.GitHash)
			url.RawQuery = query.Encode()

			return url.String(), nil
		},
	}
}

// GET / - redirects to the current month page
//
func (h Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	year, month, _ := time.Now().Date()

	reqid.FromContextToLogger(r.Context(), h.log).
		WithFields(logrus.Fields{"year": year, "month": int(month)}).
		Debug("redirect to the current month")

	url := fmt.Sprintf("/overview/%d/%d", year, month)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// GET /overview
//
func (h Handlers) OverviewPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	resp := struct {
		Footer FooterTemplateData
	}{
		Footer: FooterTemplateData{
			Version: version.Version,
			GitHash: version.GitHash,
		},
	}
	if err := h.tplExecutor.Execute(ctx, w, overviewTemplateName, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
	}
}

// GET /overview/{year:[0-9]+}
//
func (h Handlers) YearPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	year, ok := getYear(r)
	if !ok {
		h.processErrorWithPage(ctx, log, w, invalidURLMessagePrefix+"invalid year", http.StatusBadRequest)
		return
	}

	months, err := h.db.GetMonths(ctx, year)
	// Render the page even theare no months for passed year
	if err != nil && !errors.Is(err, db.ErrYearNotExist) {
		msg := dbErrorMessagePrefix + "couldn't get months"
		h.processInternalErrorWithPage(ctx, log, w, msg, err)
		return
	}

	// Display all months. Months without data in DB have zero id

	allMonths := make([]db.Month, 12)
	for month := time.January; month <= time.December; month++ {
		allMonths[month-1] = db.Month{
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
		Year     int
		NextYear int
		PrevYear int
		//
		Months       []db.Month
		AnnualIncome money.Money
		AnnualSpend  money.Money
		Result       money.Money
		//
		Footer FooterTemplateData
	}{
		Year:     year,
		NextYear: year + 1,
		PrevYear: year - 1,
		//
		Months:       allMonths,
		AnnualIncome: annualIncome,
		AnnualSpend:  annualSpend,
		Result:       result,
		//
		Footer: FooterTemplateData{
			Version: version.Version,
			GitHash: version.GitHash,
		},
	}
	if err := h.tplExecutor.Execute(ctx, w, yearTemplateName, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
	}
}

// GET /overview/{year:[0-9]+}/{month:[0-9]+}
//
//nolint:funlen
func (h Handlers) MonthPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	year, ok := getYear(r)
	if !ok {
		h.processErrorWithPage(ctx, log, w, invalidURLMessagePrefix+"invalid year", http.StatusBadRequest)
		return
	}
	monthNumber, ok := getMonth(r)
	if !ok {
		h.processErrorWithPage(ctx, log, w, invalidURLMessagePrefix+"invalid month", http.StatusBadRequest)
		return
	}

	monthID, err := h.db.GetMonthID(ctx, year, int(monthNumber))
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			h.processErrorWithPage(ctx, log, w, err.Error(), http.StatusNotFound)
		default:
			msg := dbErrorMessagePrefix + "couldn't get month"
			h.processInternalErrorWithPage(ctx, log, w, msg, err)
		}
		return
	}

	// Process
	month, err := h.db.GetMonth(ctx, monthID)
	if err != nil {
		// Month must exist
		msg := dbErrorMessagePrefix + "couldn't get Month info"
		h.processInternalErrorWithPage(ctx, log, w, msg, err)
		return
	}

	spendTypes, err := h.getSpendTypesWithFullNames(ctx)
	if err != nil {
		msg := dbErrorMessagePrefix + "couldn't get Spend Types"
		h.processInternalErrorWithPage(ctx, log, w, msg, err)
		return
	}

	sort.Slice(spendTypes, func(i, j int) bool {
		return spendTypes[i].FullName < spendTypes[j].FullName
	})

	populateMonthlyPaymentsWithFullSpendTypeNames(spendTypes, month.MonthlyPayments)
	for i := range month.Days {
		populateSpendsWithFullSpendTypeNames(spendTypes, month.Days[i].Spends)
	}

	resp := struct {
		db.Month
		SpendTypes []SpendType
		//
		Footer FooterTemplateData
		//
		ToShortMonth           func(time.Month) string
		SumSpendCosts          func([]db.Spend) money.Money
		ToHTMLAttr             func(string) template.HTMLAttr
		ShouldSuggestSpendType func(spendType, option SpendType) bool
	}{
		Month:      month,
		SpendTypes: spendTypes,
		//
		Footer: FooterTemplateData{
			Version: version.Version,
			GitHash: version.GitHash,
		},
		//
		ToShortMonth:  toShortMonth,
		SumSpendCosts: sumSpendCosts,
		ToHTMLAttr: func(s string) template.HTMLAttr {
			return template.HTMLAttr(s) //nolint:gosec
		},
		ShouldSuggestSpendType: func(origin, suggestion SpendType) bool {
			if origin.ID == suggestion.ID {
				return false
			}
			if _, ok := suggestion.parentSpendTypeIDs[origin.ID]; ok {
				return false
			}
			return true
		},
	}
	if err := h.tplExecutor.Execute(ctx, w, monthTemplateName, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
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
//   - type_id - Spend Type id to search (can be passed multiple times: ?type_id=56&type_id=58).
//               Use id '0' to search for Spends without type
//   - sort - sort type: 'title', 'date' or 'cost'
//   - order - sort order: 'asc' or 'desc'
//
func (h Handlers) SearchSpendsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	args := parseSearchSpendsArgs(r, log)
	spends, err := h.db.SearchSpends(ctx, args)
	if err != nil {
		msg := dbErrorMessagePrefix + "couldn't complete Spend search"
		h.processInternalErrorWithPage(ctx, log, w, msg, err)
		return
	}

	spendTypes, err := h.getSpendTypesWithFullNames(ctx)
	if err != nil {
		msg := dbErrorMessagePrefix + "couldn't get Spend Types"
		h.processInternalErrorWithPage(ctx, log, w, msg, err)
		return
	}
	sort.Slice(spendTypes, func(i, j int) bool {
		return spendTypes[i].FullName < spendTypes[j].FullName
	})

	populateSpendsWithFullSpendTypeNames(spendTypes, spends)

	// Execute the template
	resp := struct {
		Spends     []db.Spend
		SpendTypes []SpendType
		TotalCost  money.Money
		//
		Footer FooterTemplateData
	}{
		Spends:     spends,
		SpendTypes: spendTypes,
		TotalCost:  sumSpendCosts(spends),
		//
		Footer: FooterTemplateData{
			Version: version.Version,
			GitHash: version.GitHash,
		},
	}
	if err := h.tplExecutor.Execute(ctx, w, searchSpendsTemplateName, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
	}
}

//nolint:funlen
func parseSearchSpendsArgs(r *http.Request, log logrus.FieldLogger) db.SearchSpendsArgs {
	// Title and Notes
	title := strings.ToLower(strings.TrimSpace(r.FormValue("title")))
	notes := strings.ToLower(strings.TrimSpace(r.FormValue("notes")))

	// Min and Max Costs
	parseCost := func(paramName string) money.Money {
		costParam := r.FormValue(paramName)
		if costParam == "" {
			return 0
		}

		cost, err := strconv.ParseFloat(costParam, 64)
		if err != nil {
			log.WithError(err).WithField(paramName, costParam).Warnf("couldn't parse '%s' param", paramName)
			cost = 0
		}
		return money.FromFloat(cost)
	}
	minCost := parseCost("min_cost")
	maxCost := parseCost("max_cost")

	// After and Before
	parseTime := func(paramName string) time.Time {
		const timeLayout = "2006-01-02"

		timeParam := r.FormValue(paramName)
		if timeParam == "" {
			return time.Time{}
		}

		t, err := time.Parse(timeLayout, timeParam)
		if err != nil {
			log.WithError(err).WithField(paramName, timeParam).Warnf("couldn't parse '%s' param", paramName)
			t = time.Time{}
		}
		return t
	}
	after := parseTime("after")
	before := parseTime("before")

	// Spend Types
	var typeIDs []uint
	if ids := r.Form["type_id"]; len(ids) != 0 {
		typeIDs = make([]uint, 0, len(ids))
		for i := range ids {
			id, err := strconv.ParseUint(ids[i], 10, 0)
			if err != nil {
				log.WithError(err).WithField("type_id", ids[i]).Warn("couldn't convert Spend Type id")
				continue
			}
			typeIDs = append(typeIDs, uint(id))
		}
	}

	// Sort
	sortType := db.SortSpendsByDate
	switch r.FormValue("sort") {
	case "title":
		sortType = db.SortSpendsByTitle
	case "cost":
		sortType = db.SortSpendsByCost
	}

	// Order
	order := db.OrderByAsc
	if r.FormValue("order") == "desc" {
		order = db.OrderByDesc
	}

	return db.SearchSpendsArgs{
		Title:   title,
		Notes:   notes,
		After:   after,
		Before:  before,
		MinCost: minCost,
		MaxCost: maxCost,
		TypeIDs: typeIDs,
		Sort:    sortType,
		Order:   order,
		//
		TitleExactly: false,
		NotesExactly: false,
	}
}

type FooterTemplateData struct {
	Version string
	GitHash string
}
