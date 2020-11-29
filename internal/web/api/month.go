package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type MonthsHandlers struct {
	db  MonthsDB
	log logrus.FieldLogger
}

type MonthsDB interface {
	GetMonth(ctx context.Context, id uint) (db.Month, error)
	GetMonthID(ctx context.Context, year, month int) (uint, error)
}

// @Summary Get Month by id
// @Tags Months
// @Router /api/months/id [get]
// @Param params query models.GetMonthByIDReq true "Month ID"
// @Produce json
// @Success 200 {object} models.GetMonthResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h MonthsHandlers) GetMonthByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.GetMonthByIDReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("month_id", req.ID)

	// Process
	log.Debug("get month from the database")
	month, err := h.db.GetMonth(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			msg := "couldn't get Month with passed id"
			utils.ProcessInternalError(ctx, log, w, msg, err)
		}
		return
	}

	// Encode
	resp := models.GetMonthResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		Month: month,
	}
	utils.EncodeResponse(w, r, log, resp)
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

	log = log.WithFields(logrus.Fields{"year": req.Year, "month": req.Month})

	// Process
	log.Debug("try to get month id")
	monthID, err := h.db.GetMonthID(ctx, req.Year, req.Month)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			msg := "couldn't get month with passed year and month"
			utils.ProcessInternalError(ctx, log, w, msg, err)
		}
		return
	}

	log.Debug("get month from the database")
	month, err := h.db.GetMonth(ctx, monthID)
	if err != nil {
		msg := "couldn't get Month with passed id"
		utils.ProcessInternalError(ctx, log, w, msg, err)
		return
	}

	// Encode
	resp := models.GetMonthResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		Month: month,
	}
	utils.EncodeResponse(w, r, log, resp)
}
