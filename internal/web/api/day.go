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

type DaysHandlers struct {
	db  DaysDB
	log logrus.FieldLogger
}

type DaysDB interface {
	GetDay(ctx context.Context, id uint) (db.Day, error)
	GetDayIDByDate(ctx context.Context, year, month, day int) (id uint, err error)
}

// @Summary Get Day by id
// @Tags Days
// @Router /api/days/id [get]
// @Param params query models.GetDayByIDReq true "Day id"
// @Produce json
// @Success 200 {object} models.GetDayResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Day doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h DaysHandlers) GetDayByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.GetDayByIDReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("day_id", req.ID)

	// Process
	log.Debug("get day from the database")
	day, err := h.db.GetDay(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrDayNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			msg := "couldn't get Day with passed id"
			utils.ProcessInternalError(ctx, log, w, msg, err)
		}
		return
	}

	// Encode
	resp := models.GetDayResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		Day: day,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Get Day by date
// @Tags Days
// @Router /api/days/date [get]
// @Param params query models.GetDayByDateReq true "Date"
// @Produce json
// @Success 200 {object} models.GetDayResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Day doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h DaysHandlers) GetDayByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.GetDayByDateReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{"year": req.Year, "month": req.Month, "day": req.Day})

	// Process
	log.Debug("try to get day id")
	dayID, err := h.db.GetDayIDByDate(ctx, req.Year, req.Month, req.Day)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrDayNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			msg := "couldn't get such Day"
			utils.ProcessInternalError(ctx, log, w, msg, err)
		}
		return
	}

	log.Debug("get day from the database")
	day, err := h.db.GetDay(ctx, dayID)
	if err != nil {
		msg := "couldn't get Day with passed id"
		utils.ProcessInternalError(ctx, log, w, msg, err)
		return
	}

	// Encode
	resp := models.GetDayResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		Day: day,
	}
	utils.EncodeResponse(w, r, log, resp)
}
