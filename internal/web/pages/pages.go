package pages

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages/statistics"
)

const (
	monthsTemplateName       = "months.html"
	monthTemplateName        = "month.html"
	searchSpendsTemplateName = "search_spends.html"
	errorPageTemplateName    = "error_page.html"
)

type Handlers struct {
	db          DB
	tplExecutor *templateExecutor
	log         logrus.FieldLogger

	version string
	gitHash string
}

type DB interface {
	GetMonth(ctx context.Context, id uint) (db.Month, error)
	GetMonthID(ctx context.Context, year, month int) (uint, error)
	GetMonths(ctx context.Context, years ...int) ([]db.Month, error)

	GetSpendTypes(ctx context.Context) ([]db.SpendType, error)

	SearchSpends(ctx context.Context, args db.SearchSpendsArgs) ([]db.Spend, error)
}

func NewHandlers(db DB, log logrus.FieldLogger, cacheTemplates bool, version, gitHash string) *Handlers {
	return &Handlers{
		db:          db,
		tplExecutor: newTemplateExecutor(log, cacheTemplates, commonTemplateFuncs()),
		log:         log,
		//
		version: version,
		gitHash: gitHash,
	}
}

func commonTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"asStaticURL": func(url string) (string, error) {
			return url, nil
		},
		"toHTMLAttr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s) //nolint:gosec
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

	url := fmt.Sprintf("/%d-%d", year, month)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// GET /months?offset=0
//
func (h Handlers) MonthsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	var offset int
	if value := r.FormValue("offset"); value != "" {
		var err error
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			h.processErrorWithPage(ctx, log, w, newInvalidURLMessage("invalid offset value"), http.StatusBadRequest)
			return
		}
	}

	now := time.Now()
	endYear := now.Year() - offset
	years := []int{endYear}
	if now.Month() != time.December {
		years = append(years, endYear-1)
	}
	months, err := h.db.GetMonths(ctx, years...)
	if err != nil {
		h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't get months"), err)
		return
	}

	months = getLastTwelveMonths(endYear, now.Month(), months)

	var totalIncome money.Money
	for _, m := range months {
		totalIncome = totalIncome.Add(m.TotalIncome)
	}

	var totalSpend money.Money
	for _, m := range months {
		// Use Add because 'TotalSpend' is negative
		totalSpend = totalSpend.Add(m.TotalSpend)
	}

	// Use Add because 'annualSpend' is negative
	result := totalIncome.Add(totalSpend)

	yearInterval := strconv.Itoa(endYear)
	if len(years) > 1 {
		yearInterval = strconv.Itoa(endYear-1) + "–" + yearInterval
	}

	resp := struct {
		YearInterval string
		Offset       int
		//
		Months      []db.Month
		TotalIncome money.Money
		TotalSpend  money.Money
		Result      money.Money
		//
		Footer FooterTemplateData
		//
		Add func(int, int) int
	}{
		YearInterval: yearInterval,
		Offset:       offset,
		//
		Months:      months,
		TotalIncome: totalIncome,
		TotalSpend:  totalSpend,
		Result:      result,
		//
		Footer: FooterTemplateData{
			Version: h.version,
			GitHash: h.gitHash,
		},
		//
		Add: func(a, b int) int { return a + b },
	}
	if err := h.tplExecutor.Execute(ctx, w, monthsTemplateName, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
	}
}

// getLastTwelveMonths returns the last 12 months according to the passed year and month. If some month
// can't be found in the passed slice, its id will be 0
func getLastTwelveMonths(endYear int, endMonth time.Month, months []db.Month) []db.Month {
	type key struct {
		year  int
		month time.Month
	}
	requiredMonths := make(map[key]db.Month)

	year, month := endYear, endMonth
	for i := 0; i < 12; i++ {
		// Months without data have zero id
		requiredMonths[key{year, month}] = db.Month{ID: 0, Year: year, Month: month}

		month--
		if month == 0 {
			month = time.December
			year--
		}
	}

	for _, m := range months {
		k := key{m.Year, m.Month}
		if _, ok := requiredMonths[k]; ok {
			requiredMonths[k] = m
		}
	}

	months = make([]db.Month, 0, len(requiredMonths))
	for _, m := range requiredMonths {
		months = append(months, m)
	}
	sort.Slice(months, func(i, j int) bool {
		if months[i].Year == months[j].Year {
			return months[i].Month < months[j].Month
		}
		return months[i].Year < months[j].Year
	})

	return months
}

// GET /{year:[0-9]+}-{month:[0-9]+}
//
//nolint:funlen
func (h Handlers) MonthPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	year, ok := getYear(r)
	if !ok {
		h.processErrorWithPage(ctx, log, w, newInvalidURLMessage("invalid year"), http.StatusBadRequest)
		return
	}
	monthNumber, ok := getMonth(r)
	if !ok {
		h.processErrorWithPage(ctx, log, w, newInvalidURLMessage("invalid month"), http.StatusBadRequest)
		return
	}

	monthID, err := h.db.GetMonthID(ctx, year, int(monthNumber))
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			h.processErrorWithPage(ctx, log, w, err.Error(), http.StatusNotFound)
		default:
			h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't get month"), err)
		}
		return
	}

	// Process
	month, err := h.db.GetMonth(ctx, monthID)
	if err != nil {
		// Month must exist
		h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't get Month info"), err)
		return
	}

	dbSpendTypes, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't get Spend Types"), err)
		return
	}
	spendTypes := getSpendTypesWithFullNames(dbSpendTypes)

	populateMonthlyPaymentsWithFullSpendTypeNames(spendTypes, month.MonthlyPayments)
	for i := range month.Days {
		populateSpendsWithFullSpendTypeNames(spendTypes, month.Days[i].Spends)
	}

	// Sort Incomes and Monthly Payments
	sort.Slice(month.Incomes, func(i, j int) bool {
		return month.Incomes[i].Income > month.Incomes[j].Income
	})
	sort.Slice(month.MonthlyPayments, func(i, j int) bool {
		return month.MonthlyPayments[i].Cost > month.MonthlyPayments[j].Cost
	})

	var monthlyPaymentsTotalCost money.Money
	for _, p := range month.MonthlyPayments {
		monthlyPaymentsTotalCost = monthlyPaymentsTotalCost.Sub(p.Cost)
	}

	resp := struct {
		db.Month
		MonthlyPaymentsTotalCost money.Money
		SpendTypes               []SpendType
		//
		Footer FooterTemplateData
		//
		ToShortMonth           func(time.Month) string
		SumSpendCosts          func([]db.Spend) money.Money
		ShouldSuggestSpendType func(spendType, option SpendType) bool
	}{
		Month:                    month,
		MonthlyPaymentsTotalCost: monthlyPaymentsTotalCost,
		SpendTypes:               spendTypes,
		//
		Footer: FooterTemplateData{
			Version: h.version,
			GitHash: h.gitHash,
		},
		//
		ToShortMonth:  toShortMonth,
		SumSpendCosts: sumSpendCosts,
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
		h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't complete Spend search"), err)
		return
	}

	dbSpendTypes, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		h.processInternalErrorWithPage(ctx, log, w, newDBErrorMessage("couldn't get Spend Types"), err)
		return
	}
	spendTypes := getSpendTypesWithFullNames(dbSpendTypes)

	populateSpendsWithFullSpendTypeNames(spendTypes, spends)

	spentBySpendTypeDatasets := statistics.CalculateSpentBySpendType(dbSpendTypes, spends)
	spentByDayDataset := statistics.CalculateSpentByDay(spends, args.After, args.Before)
	// TODO: support custom interval number?
	const costIntervalNumber = 15
	costIntervals := statistics.CalculateCostIntervals(spends, costIntervalNumber)

	// Execute the template
	resp := struct {
		// Spends
		Spends []db.Spend
		// Statistics
		SpentBySpendTypeDatasets []statistics.SpentBySpendTypeDataset
		SpentByDayDataset        statistics.SpentByDayDataset
		CostIntervals            []statistics.CostInterval
		TotalCost                money.Money
		//
		SpendTypes []SpendType
		Footer     FooterTemplateData
	}{
		Spends: spends,
		//
		SpentBySpendTypeDatasets: spentBySpendTypeDatasets,
		SpentByDayDataset:        spentByDayDataset,
		CostIntervals:            costIntervals,
		TotalCost:                sumSpendCosts(spends),
		//
		SpendTypes: spendTypes,
		Footer: FooterTemplateData{
			Version: h.version,
			GitHash: h.gitHash,
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

const (
	executeErrorMessage = "Can't execute template"
)

func newInvalidURLMessage(err string) string {
	return "Invalid URL: " + err
}

func newDBErrorMessage(err string) string {
	return "DB error: " + err
}
