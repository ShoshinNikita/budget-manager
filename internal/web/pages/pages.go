package pages

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/version"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages/templates"
)

const (
	executeErrorMessage     = "Can't execute template"
	invalidURLMessagePrefix = "Invalid URL: "
	dbErrorMessagePrefix    = "DB error: "
)

//nolint:gochecknoglobals
var (
	overviewTemplatePath = templates.Template{
		Path: "templates/overview.html",
		Deps: []string{"templates/footer.html"},
	}
	yearTemplatePath = templates.Template{
		Path: "templates/overview_year.html",
		Deps: []string{"templates/footer.html"},
	}
	monthTemplatePath = templates.Template{
		Path: "templates/overview_year_month.html",
		Deps: []string{"templates/footer.html"},
	}
	//
	searchSpendsTemplatePath = templates.Template{
		Path: "templates/search_spends.html",
		Deps: []string{"templates/footer.html"},
	}
	//
	errorPageTemplatePath = templates.Template{Path: "./templates/error_page.html"}
)

type Handlers struct {
	db          DB
	tplExecutor TemplateExecutor
	log         logrus.FieldLogger
}

type DB interface {
	GetMonth(ctx context.Context, id uint) (db.Month, error)
	GetMonthID(ctx context.Context, year, month int) (uint, error)
	GetMonths(ctx context.Context, year int) ([]db.Month, error)

	GetSpendTypes(ctx context.Context) ([]db.SpendType, error)

	SearchSpends(ctx context.Context, args db.SearchSpendsArgs) ([]db.Spend, error)
}

type TemplateExecutor interface {
	Execute(ctx context.Context, t templates.Template, w io.Writer, data interface{}) error
}

func NewHandlers(db DB, tplExecutor TemplateExecutor, log logrus.FieldLogger) *Handlers {
	return &Handlers{
		db:          db,
		tplExecutor: tplExecutor,
		log:         log,
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

	if err := h.tplExecutor.Execute(ctx, overviewTemplatePath, w, nil); err != nil {
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
		Year         int
		Months       []db.Month
		AnnualIncome money.Money
		AnnualSpend  money.Money
		Result       money.Money
		//
		Footer FooterTemplateData
	}{
		Year:         year,
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
	if err := h.tplExecutor.Execute(ctx, yearTemplatePath, w, resp); err != nil {
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
	if err := h.tplExecutor.Execute(ctx, monthTemplatePath, w, resp); err != nil {
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
//nolint:funlen
func (h Handlers) SearchSpendsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Parse the query

	// Parse Title and Notes
	title := strings.TrimSpace(r.FormValue("title"))
	notes := strings.TrimSpace(r.FormValue("notes"))

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

	// Parse Spend Types
	var typeIDs []uint
	if ids := r.Form["type_id"]; len(ids) != 0 {
		typeIDs = make([]uint, 0, len(ids))
		for i := range ids {
			id, err := strconv.ParseUint(ids[i], 10, 0)
			if err != nil {
				// Just log the error
				log.WithError(err).WithField("type_id", ids[i]).Warn("couldn't convert Spend Type id")
				continue
			}
			typeIDs = append(typeIDs, uint(id))
		}
	}

	// Sort
	sortType := func() db.SearchSpendsColumn {
		switch r.FormValue("sort") {
		case "title":
			return db.SortSpendsByTitle
		case "cost":
			return db.SortSpendsByCost
		default:
			return db.SortSpendsByDate
		}
	}()
	order := func() db.SearchOrder {
		switch r.FormValue("order") {
		case "desc":
			return db.OrderByDesc
		default:
			return db.OrderByAsc
		}
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
		Sort:    sortType,
		Order:   order,
		// TODO
		TitleExactly: false,
		NotesExactly: false,
	}
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
	if err := h.tplExecutor.Execute(ctx, searchSpendsTemplatePath, w, resp); err != nil {
		h.processInternalErrorWithPage(ctx, log, w, executeErrorMessage, err)
	}
}

type FooterTemplateData struct {
	Version string
	GitHash string
}
