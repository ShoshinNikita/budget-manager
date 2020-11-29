package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SpendsHandlers struct {
	db  SpendsDB
	log logrus.FieldLogger
}

type SpendsDB interface {
	AddSpend(ctx context.Context, args db.AddSpendArgs) (id uint, err error)
	EditSpend(ctx context.Context, args db.EditSpendArgs) error
	RemoveSpend(ctx context.Context, id uint) error
}

// @Summary Create Spend
// @Tags Spends
// @Router /api/spends [post]
// @Accept json
// @Param body body models.AddSpendReq true "New Spend"
// @Produce json
// @Success 200 {object} models.AddSpendResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Day doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendsHandlers) AddSpend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.AddSpendReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{
		"day_id": req.DayID, "title": req.Title, "type_id": req.TypeID,
		"notes": req.Notes, "cost": req.Cost,
	})

	// Process
	log.Debug("add Spend")
	args := db.AddSpendArgs{
		DayID:  req.DayID,
		Title:  strings.TrimSpace(req.Title),
		TypeID: req.TypeID,
		Notes:  strings.TrimSpace(req.Notes),
		Cost:   money.FromFloat(req.Cost),
	}
	id, err := h.db.AddSpend(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrDayNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't add Spend", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Spend was successfully added")

	// Encode
	resp := models.AddSpendResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		ID: id,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Edit Spend
// @Tags Spends
// @Router /api/spends [put]
// @Accept json
// @Param body body models.EditSpendReq true "Updated Spend"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendsHandlers) EditSpend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.EditSpendReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "type_id": req.TypeID,
	})

	// Process
	log.Debug("edit Spend")
	args := db.EditSpendArgs{
		ID:     req.ID,
		Title:  trimSpacePointer(req.Title),
		Notes:  trimSpacePointer(req.Notes),
		TypeID: req.TypeID,
	}
	if req.Cost != nil {
		cost := money.FromFloat(*req.Cost)
		args.Cost = &cost
	}
	err := h.db.EditSpend(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't edit Spend", err)
		}
		return
	}
	log.Info("Spend was successfully edited")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Remove Spend
// @Tags Spends
// @Router /api/spends [delete]
// @Accept json
// @Param body body models.RemoveSpendReq true "Updated Spend"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendsHandlers) RemoveSpend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.RemoveSpendReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Spend")
	err := h.db.RemoveSpend(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't remove Spend", err)
		}
		return
	}
	log.Info("Spend was successfully removed")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}
