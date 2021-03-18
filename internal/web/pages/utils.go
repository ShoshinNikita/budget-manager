package pages

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

// processErrorWithPage is similar to 'utils.ProcessError' but shows the error page instead of returning json
//
//nolint:gofumpt
func (h Handlers) processErrorWithPage(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int) {

	data := struct {
		Code      int
		RequestID reqid.RequestID
		Message   string
		//
		Footer FooterTemplateData
	}{
		Code:      code,
		RequestID: reqid.FromContext(ctx),
		Message:   respMsg,
		//
		Footer: FooterTemplateData{
			Version: h.version,
			GitHash: h.gitHash,
		},
	}
	if err := h.tplExecutor.Execute(ctx, w, errorPageTemplateName, data); err != nil {
		utils.ProcessInternalError(ctx, log, w, executeErrorMessage, err)
	}
}

// processInternalErrorWithPage is similar to 'utils.ProcessInternalError' but shows the error page
// instead of returning json
//
//nolint:gofumpt
func (h Handlers) processInternalErrorWithPage(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, err error) {

	utils.LogInternalError(log, respMsg, err)

	h.processErrorWithPage(ctx, log, w, respMsg, http.StatusInternalServerError)
}

func toShortMonth(m time.Month) string {
	month := m.String()
	// Don't trim June and July
	if len(month) > 4 {
		month = m.String()[:3]
	}
	return month
}

func sumSpendCosts(spends []db.Spend) money.Money {
	var m money.Money
	for i := range spends {
		m = m.Sub(spends[i].Cost)
	}
	return m
}

const yearKey = "year"

func getYear(r *http.Request) (year int, ok bool) {
	h, ok := mux.Vars(r)[yearKey]
	if !ok {
		return 0, false
	}

	year, err := strconv.Atoi(h)
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
