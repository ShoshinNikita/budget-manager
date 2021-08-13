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

type IncomesHandlers struct {
	db  IncomesDB
	log logger.Logger
}

type IncomesDB interface {
	AddIncome(ctx context.Context, args db.AddIncomeArgs) (id uint, err error)
	EditIncome(ctx context.Context, args db.EditIncomeArgs) error
	RemoveIncome(ctx context.Context, id uint) error
}

// @Summary Create Income
// @Tags Incomes
// @Router /api/incomes [post]
// @Accept json
// @Param body body models.AddIncomeReq true "New Income"
// @Produce json
// @Success 201 {object} models.AddIncomeResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h IncomesHandlers) AddIncome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.AddIncomeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}
	log = log.WithRequest(req)

	// Process
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		Notes:   req.Notes,
		Income:  money.FromFloat(req.Income),
	}
	id, err := h.db.AddIncome(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't add Income", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Debug("Income was successfully added")

	resp := &models.AddIncomeResp{
		ID: id,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp), utils.EncodeStatusCode(http.StatusCreated))
}

// @Summary Edit Income
// @Tags Incomes
// @Router /api/incomes [put]
// @Accept json
// @Param body body models.EditIncomeReq true "Updated Income"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Income doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h IncomesHandlers) EditIncome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.EditIncomeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}
	log = log.WithRequest(req)

	// Process
	args := db.EditIncomeArgs{
		ID:    req.ID,
		Title: req.Title,
		Notes: req.Notes,
	}
	if req.Income != nil {
		income := money.FromFloat(*req.Income)
		args.Income = &income
	}
	err := h.db.EditIncome(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrIncomeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't edit Income", err)
		}
		return
	}
	log.Debug("Income was successfully edited")

	utils.Encode(ctx, w, log)
}

// @Summary Remove Income
// @Tags Incomes
// @Router /api/incomes [delete]
// @Accept json
// @Param body body models.RemoveIncomeReq true "Income id"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Income doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h IncomesHandlers) RemoveIncome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.RemoveIncomeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}
	log = log.WithRequest(req)

	// Process
	err := h.db.RemoveIncome(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrIncomeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't remove Income", err)
		}
		return
	}
	log.Debug("Income was successfully removed")

	utils.Encode(ctx, w, log)
}
