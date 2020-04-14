package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/models"
)

const (
	errDecodeRequest  = "couldn't decode request"
	errEncodeResponse = "couldn't encode response"
)

// GET / - redirects to the current month page
//
func (s Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	year, month, _ := time.Now().Date()
	log = log.WithFields(logrus.Fields{"year": year, "month": int(month)})

	log.Debug("redirect to the current month")
	url := fmt.Sprintf("/overview/%d/%d", year, month)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// @Summary Get Month by id
// @Tags Months
// @Router /api/months/id [get]
// @Accept json
// @Param body body models.GetMonthByIDReq true "Month ID"
// @Produce json
// @Success 200 {object} models.GetMonthResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) GetMonthByID(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.GetMonthByIDReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("month_id", req.ID)

	// Process
	log.Debug("get month from the database")
	month, err := s.db.GetMonth(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get Month with passed id"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.GetMonthResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		Month: *month,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Get Month by date
// @Tags Months
// @Router /api/months/date [get]
// @Accept json
// @Param body body models.GetMonthByDateReq true "Date"
// @Produce json
// @Success 200 {object} models.GetMonthResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Month doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) GetMonthByDate(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.GetMonthByDateReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{"year": req.Year, "month": req.Month})

	log.Debug("try to get month id")
	monthID, err := s.db.GetMonthID(r.Context(), req.Year, req.Month)
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get month with passed year and month"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return
	}

	// Process
	log.Debug("get month from the database")
	month, err := s.db.GetMonth(r.Context(), monthID)
	if err != nil {
		msg := "couldn't get Month with passed id"
		s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		return
	}

	resp := models.GetMonthResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		Month: *month,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Get Day by id
// @Tags Days
// @Router /api/days/id [get]
// @Accept json
// @Param body body models.GetDayByIDReq true "Day id"
// @Produce json
// @Success 200 {object} models.GetDayResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Day doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) GetDayByID(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.GetDayByIDReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("day_id", req.ID)

	// Process
	log.Debug("get day from the database")
	day, err := s.db.GetDay(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrDayNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get Day with passed id"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.GetDayResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		Day: *day,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Get Day by date
// @Tags Days
// @Router /api/days/date [get]
// @Accept json
// @Param body body models.GetDayByDateReq true "Date"
// @Produce json
// @Success 200 {object} models.GetDayResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Day doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) GetDayByDate(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.GetDayByDateReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{"year": req.Year, "month": req.Month, "day": req.Day})

	// Prepare
	log.Debug("try to get day id")
	dayID, err := s.db.GetDayIDByDate(r.Context(), req.Year, req.Month, req.Day)
	if err != nil {
		switch err {
		case db.ErrDayNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get such Day"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return
	}

	// Process
	log.Debug("get day from the database")
	day, err := s.db.GetDay(r.Context(), dayID)
	if err != nil {
		msg := "couldn't get Day with passed id"
		s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		return
	}

	resp := models.GetDayResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		Day: *day,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Income
// -------------------------------------------------

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
func (s Server) AddIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.AddIncomeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"month_id": req.MonthID, "title": req.Title, "notes": req.Notes, "income": req.Income,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Process
	log.Debug("add Income")
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   strings.TrimSpace(req.Title),
		Notes:   strings.TrimSpace(req.Notes),
		Income:  money.FromFloat(req.Income),
	}
	id, err := s.db.AddIncome(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't add Income", http.StatusInternalServerError, err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Income was successfully added")

	resp := models.AddIncomeResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) EditIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.EditIncomeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "income": req.Income,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

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
	err := s.db.EditIncome(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrIncomeNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't edit Income", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Income was successfully edited")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) RemoveIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.RemoveIncomeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Income")
	err := s.db.RemoveIncome(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrIncomeNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't remove Income", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Income was successfully removed")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Monthly Payment
// -------------------------------------------------

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
func (s Server) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.AddMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"month_id": req.MonthID, "title": req.Title, "type_id": req.TypeID,
		"notes": req.Notes, "cost": req.Cost,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Process
	log.Debug("add Monthly Payment")
	args := db.AddMonthlyPaymentArgs{
		MonthID: req.MonthID,
		Title:   strings.TrimSpace(req.Title),
		TypeID:  req.TypeID,
		Notes:   strings.TrimSpace(req.Notes),
		Cost:    money.FromFloat(req.Cost),
	}
	id, err := s.db.AddMonthlyPayment(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't add Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Monthly Payment was successfully added")

	resp := models.AddMonthlyPaymentResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.EditMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "type_id": req.TypeID, "cost": req.Cost,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

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
	err := s.db.EditMonthlyPayment(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrMonthlyPaymentNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't edit Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Monthly Payment was successfully edited")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) RemoveMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.RemoveMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Monthly Payment")
	err := s.db.RemoveMonthlyPayment(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrMonthlyPaymentNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't remove Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Monthly Payment was successfully removed")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Spend
// -------------------------------------------------

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
func (s Server) AddSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.AddSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"day_id": req.DayID, "title": req.Title, "type_id": req.TypeID,
		"notes": req.Notes, "cost": req.Cost,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Process
	log.Debug("add Spend")
	args := db.AddSpendArgs{
		DayID:  req.DayID,
		Title:  strings.TrimSpace(req.Title),
		TypeID: req.TypeID,
		Notes:  strings.TrimSpace(req.Notes),
		Cost:   money.FromFloat(req.Cost),
	}
	id, err := s.db.AddSpend(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrDayNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't add Spend", http.StatusInternalServerError, err)
		}
		return
	}
	log = log.WithField("id", id)
	log.Info("Spend was successfully added")

	resp := models.AddSpendResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) EditSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.EditSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"id": req.ID, "title": req.Title, "notes": req.Notes, "type_id": req.TypeID,
	})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

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
	err := s.db.EditSpend(r.Context(), args)
	if err != nil {
		switch err {
		case db.ErrSpendNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't edit Spend", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend was successfully edited")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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
func (s Server) RemoveSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.RemoveSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Spend")
	err := s.db.RemoveSpend(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrSpendNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't remove Spend", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend was successfully removed")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Spend Types
// -------------------------------------------------

// @Summary Get All Spend Types
// @Tags Spend Types
// @Router /api/spend-types [get]
// @Produce json
// @Success 200 {object} models.GetSpendTypesResp
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) GetSpendTypes(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Process
	log.Debug("return all Spend Types")
	types, err := s.db.GetSpendTypes(r.Context())
	if err != nil {
		s.processError(r.Context(), log, w, "couldn't get Spend Types", http.StatusInternalServerError, err)
		return
	}

	resp := models.GetSpendTypesResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		SpendTypes: types,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Create Spend Type
// @Tags Spend Types
// @Router /api/spend-types [post]
// @Accept json
// @Param body body models.AddSpendTypeReq true "New Spend Type"
// @Produce json
// @Success 200 {object} models.AddSpendTypeResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) AddSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.AddSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("name", req.Name)

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Process
	log.Debug("add Spend Type")
	id, err := s.db.AddSpendType(r.Context(), strings.TrimSpace(req.Name))
	if err != nil {
		s.processError(r.Context(), log, w, "couldn't add Spend Type", http.StatusInternalServerError, err)
		return
	}
	log = log.WithField("id", id)
	log.Info("Spend Type was successfully added")

	resp := models.AddSpendTypeResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Edit Spend Type
// @Tags Spend Types
// @Router /api/spend-types [put]
// @Accept json
// @Param body body models.EditSpendTypeReq true "Updated Spend Type"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend Type doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) EditSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.EditSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{"id": req.ID, "name": req.Name})

	// Check request
	if err := req.Check(); err != nil {
		s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Process
	log.Debug("edit Spend Type")
	err := s.db.EditSpendType(r.Context(), req.ID, strings.TrimSpace(req.Name))
	if err != nil {
		switch err {
		case db.ErrSpendTypeNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't edit Spend Type", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend Type was successfully edited")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// @Summary Remove Spend Type
// @Tags Spend Types
// @Router /api/spend-types [delete]
// @Accept json
// @Param body body models.RemoveSpendTypeReq true "Spend Type id"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend Type doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) RemoveSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.RemoveSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Spend Type")
	err := s.db.RemoveSpendType(r.Context(), req.ID)
	if err != nil {
		switch err {
		case db.ErrSpendTypeNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, err)
		default:
			s.processError(r.Context(), log, w, "couldn't remove Spend Type", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend Type was successfully removed")

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Other
// -------------------------------------------------

// @Summary Search Spends
// @Tags Search
// @Router /api/search/spends [get]
// @Accept json
// @Param body body models.SearchSpendsReq true "Search args"
// @Produce json
// @Success 200 {object} models.SearchSpendsResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 500 {object} models.Response "Internal error"
//
func (s Server) SearchSpends(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	// Decode
	req := &models.SearchSpendsReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	log = log.WithFields(logrus.Fields{
		"title": req.Title, "title_exactly": req.TitleExactly,
		"notes": req.Notes, "notes_exactly": req.NotesExactly,
		"after": req.After, "before": req.Before, "type_ids": req.TypeIDs,
		"min_cost": req.MinCost, "max_cost": req.MaxCost,
		"sort": req.Sort, "order": req.Order,
	})

	// Process
	log.Debug("search for Spends")
	args := db.SearchSpendsArgs{
		Title:        strings.ToLower(strings.TrimSpace(req.Title)),
		Notes:        strings.ToLower(strings.TrimSpace(req.Notes)),
		TitleExactly: req.TitleExactly,
		NotesExactly: req.NotesExactly,
		After:        req.After,
		Before:       req.Before,
		MinCost:      money.FromFloat(req.MinCost),
		MaxCost:      money.FromFloat(req.MaxCost),
		WithoutType:  req.WithoutType,
	}
	if !args.WithoutType {
		args.TypeIDs = req.TypeIDs
	}
	switch req.Sort {
	case "title":
		args.Sort = db.SortSpendsByTitle
	case "cost":
		args.Sort = db.SortSpendsByCost
	default:
		args.Sort = db.SortSpendsByDate
	}
	switch req.Order {
	case "desc":
		args.Order = db.OrderByDesc
	default:
		args.Order = db.OrderByAsc
	}

	spends, err := s.db.SearchSpends(r.Context(), args)
	if err != nil {
		s.processError(r.Context(), log, w, "couldn't search for Spends", http.StatusInternalServerError, err)
		return
	}
	log.WithField("spend_number", len(spends)).Debug("finish Spend search")

	resp := models.SearchSpendsResp{
		Response: models.Response{
			RequestID: request_id.FromContext(r.Context()).ToString(),
			Success:   true,
		},
		Spends: spends,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), log, w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Helpers
// -------------------------------------------------

//nolint:unused,deadcode,errcheck
func notImplementedYet(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("not implemented yet"))
}

// trimSpacePointer is like 'strings.TrimPointer' but for pointers
func trimSpacePointer(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

// jsonNewDecoder is a wrapper for json.NewDecoder function.
// It creates a new json.Decoder and calls json.Decoder.DisallowUnknownFields method
func jsonNewDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	return d
}
