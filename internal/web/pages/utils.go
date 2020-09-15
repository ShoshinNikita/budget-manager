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
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

// processErrorWithPage is similar to 'processError', but it shows error page instead of returning json
//
//nolint:gofumpt
func (h Handlers) processErrorWithPage(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		utils.LogInternalError(log, respMsg, internalErr)
	}

	data := struct {
		Code      int
		RequestID request_id.RequestID
		Message   string
	}{
		Code:      code,
		RequestID: request_id.FromContext(ctx),
		Message:   respMsg,
	}
	if err := h.tplExecutor.Execute(ctx, errorPageTemplatePath, w, data); err != nil {
		utils.ProcessInternalError(ctx, log, w, executeErrorMessage, err)
	}
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
