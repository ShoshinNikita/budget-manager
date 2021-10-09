package pages

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

// processErrorWithPage is similar to 'utils.ProcessError' but shows the error page instead of returning json
func (h Handlers) processErrorWithPage(ctx context.Context, log logger.Logger, w http.ResponseWriter,
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
		utils.EncodeInternalError(ctx, w, log, executeErrorMessage, err)
	}
}

// processInternalErrorWithPage is similar to 'utils.ProcessInternalError' but shows the error page
// instead of returning json
func (h Handlers) processInternalErrorWithPage(ctx context.Context, log logger.Logger, w http.ResponseWriter,
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

func getYearAndMonth(r *http.Request) (y int, month time.Month, ok bool) {
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		return 0, 0, false
	}

	m, err := strconv.Atoi(r.FormValue("month"))
	if err != nil {
		return 0, 0, false
	}
	month = time.Month(m)
	if !(time.January <= month && month <= time.December) {
		return 0, 0, false
	}

	return year, month, true
}
