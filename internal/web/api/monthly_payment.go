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

type MonthlyPaymentsHandlers struct {
	db  MonthlyPaymentsDB
	log logger.Logger
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
// @Success 201 {object} models.AddMonthlyPaymentResp
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
	log = log.WithRequest(req)

	// Process
	args := db.AddMonthlyPaymentArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		TypeID:  req.TypeID,
		Notes:   req.Notes,
		Cost:    money.FromFloat(req.Cost),
	}
	id, err := h.db.AddMonthlyPayment(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusBadRequest)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't add Monthly Payment", err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Debug("Monthly Payment was successfully added")

	resp := &models.AddMonthlyPaymentResp{
		ID: id,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp), utils.EncodeStatusCode(http.StatusCreated))
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
	log = log.WithRequest(req)

	// Process
	args := db.EditMonthlyPaymentArgs{
		ID:     req.ID,
		Title:  req.Title,
		Notes:  req.Notes,
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
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusBadRequest)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't edit Monthly Payment", err)
		}
		return
	}
	log.Debug("Monthly Payment was successfully edited")

	utils.Encode(ctx, w, log)
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
	log = log.WithRequest(req)

	// Process
	err := h.db.RemoveMonthlyPayment(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrMonthlyPaymentNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't remove Monthly Payment", err)
		}
		return
	}
	log.Debug("Monthly Payment was successfully removed")

	utils.Encode(ctx, w, log)
}
