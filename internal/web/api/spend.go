package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SpendsHandlers struct {
	db  SpendsDB
	log logger.Logger
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
// @Success 201 {object} models.AddSpendResp
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
	log = log.WithRequest(req)

	// Process
	args := db.AddSpendArgs{
		DayID:  req.DayID,
		Title:  req.Title,
		TypeID: req.TypeID,
		Notes:  req.Notes,
		Cost:   money.FromFloat(req.Cost),
	}
	id, err := h.db.AddSpend(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrDayNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusBadRequest)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't add Spend", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Debug("Spend was successfully added")

	resp := &models.AddSpendResp{
		ID: id,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp), utils.EncodeStatusCode(http.StatusCreated))
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
	log = log.WithRequest(req)

	// Process
	args := db.EditSpendArgs{
		ID:     req.ID,
		Title:  req.Title,
		Notes:  req.Notes,
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
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusBadRequest)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't edit Spend", err)
		}
		return
	}
	log.Debug("Spend was successfully edited")

	utils.Encode(ctx, w, log)
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
	log = log.WithRequest(req)

	// Process
	err := h.db.RemoveSpend(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't remove Spend", err)
		}
		return
	}
	log.Debug("Spend was successfully removed")

	utils.Encode(ctx, w, log)
}
