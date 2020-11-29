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

type IncomesHandlers struct {
	db  IncomesDB
	log logrus.FieldLogger
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
// @Success 200 {object} models.AddIncomeResp
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

	log = log.WithFields(logrus.Fields{
		"month_id": req.MonthID, "title": req.Title, "notes": req.Notes, "income": req.Income,
	})

	// Process
	log.Debug("add Income")
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   strings.TrimSpace(req.Title),
		Notes:   strings.TrimSpace(req.Notes),
		Income:  money.FromFloat(req.Income),
	}
	id, err := h.db.AddIncome(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't add Income", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Income was successfully added")

	// Encode
	resp := models.AddIncomeResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		ID: id,
	}
	utils.EncodeResponse(w, r, log, resp)
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

	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "income": req.Income,
	})

	// Process
	log.Debug("edit Income")
	args := db.EditIncomeArgs{
		ID:    req.ID,
		Title: trimSpacePointer(req.Title),
		Notes: trimSpacePointer(req.Notes),
	}
	if req.Income != nil {
		income := money.FromFloat(*req.Income)
		args.Income = &income
	}
	err := h.db.EditIncome(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrIncomeNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't edit Income", err)
		}
		return
	}
	log.Info("Income was successfully edited")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
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

	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Income")
	err := h.db.RemoveIncome(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrIncomeNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't remove Income", err)
		}
		return
	}
	log.Info("Income was successfully removed")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}
