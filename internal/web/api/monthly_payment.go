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

type MonthlyPaymentsHandlers struct {
	db  MonthlyPaymentsDB
	log logrus.FieldLogger
}

type MonthlyPaymentsDB interface {
	AddMonthlyPayment(ctx context.Context, args db.AddMonthlyPaymentArgs) (id uint, err error)
	EditMonthlyPayment(ctx context.Context, args db.EditMonthlyPaymentArgs) error
	RemoveMonthlyPayment(ctx context.Context, id uint) error
}

// @Summary Create Monthly Payment
// @Tags Monthly Payments
// @Router /api/monthly-payments [post]
// @Accept json
// @Param body body models.AddMonthlyPaymentReq true "New Monthly Payment"
// @Produce json
// @Success 200 {object} models.AddMonthlyPaymentResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h MonthlyPaymentsHandlers) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.AddMonthlyPaymentReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{
		"month_id": req.MonthID, "title": req.Title, "type_id": req.TypeID,
		"notes": req.Notes, "cost": req.Cost,
	})

	// Process
	log.Debug("add Monthly Payment")
	args := db.AddMonthlyPaymentArgs{
		MonthID: req.MonthID,
		Title:   strings.TrimSpace(req.Title),
		TypeID:  req.TypeID,
		Notes:   strings.TrimSpace(req.Notes),
		Cost:    money.FromFloat(req.Cost),
	}
	id, err := h.db.AddMonthlyPayment(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't add Monthly Payment", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Monthly Payment was successfully added")

	// Encode
	resp := models.AddMonthlyPaymentResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		ID: id,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Edit Monthly Payment
// @Tags Monthly Payments
// @Router /api/monthly-payments [put]
// @Accept json
// @Param body body models.EditMonthlyPaymentReq true "Updated Monthly Payment"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Monthly Payment doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h MonthlyPaymentsHandlers) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.EditMonthlyPaymentReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "type_id": req.TypeID, "cost": req.Cost,
	})

	// Process
	log.Debug("edit Monthly Payment")
	args := db.EditMonthlyPaymentArgs{
		ID:     req.ID,
		Title:  trimSpacePointer(req.Title),
		Notes:  trimSpacePointer(req.Notes),
		TypeID: req.TypeID,
	}
	if req.Cost != nil {
		cost := money.FromFloat(*req.Cost)
		args.Cost = &cost
	}
	err := h.db.EditMonthlyPayment(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthlyPaymentNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't edit Monthly Payment", err)
		}
		return
	}
	log.Info("Monthly Payment was successfully edited")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Remove Monthly Payment
// @Tags Monthly Payments
// @Router /api/monthly-payments [delete]
// @Accept json
// @Param body body models.RemoveMonthlyPaymentReq true "Monthly Payment id"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Monthly Payment doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h MonthlyPaymentsHandlers) RemoveMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.RemoveMonthlyPaymentReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Monthly Payment")
	err := h.db.RemoveMonthlyPayment(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthlyPaymentNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't remove Monthly Payment", err)
		}
		return
	}
	log.Info("Monthly Payment was successfully removed")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}
