package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type MonthsHandlers struct {
	db  MonthsDB
	log logger.Logger
}

type MonthsDB interface {
	GetMonthByDate(ctx context.Context, year int, month time.Month) (db.Month, error)
}

// @Summary Get Month by date
// @Tags Months
// @Router /api/months/date [get]
// @Param params query models.GetMonthByDateReq true "Date"
// @Produce json
// @Success 200 {object} models.GetMonthResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h MonthsHandlers) GetMonthByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.GetMonthByDateReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}
	log = log.WithRequest(req)

	// Process
	month, err := h.db.GetMonthByDate(ctx, req.Year, req.Month)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't get Month for passed year and month", err)
		}
		return
	}

	resp := &models.GetMonthResp{
		Month: month,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp))
}
