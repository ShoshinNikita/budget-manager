package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// GET /api/months
//
// Request: models.GetMonthReq or models.GetMonthByYearAndMonthReq
// Response: models.GetMonthResp or models.Response
//
func (s Server) GetMonth(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

	// Prepare
	var monthID uint
	monthID, ok := s.getMonthID(w, r)
	if !ok {
		// 'Server.getMonthID' has already called 'Server.processError'
		return
	}
	log = log.WithField("month_id", monthID)

	// Process
	log.Debug("get month from the database")
	month, err := s.db.GetMonth(r.Context(), monthID)
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

func (s Server) getMonthID(w http.ResponseWriter, r *http.Request) (id uint, ok bool) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.processError(r.Context(), log, w, "couldn't read body", http.StatusBadRequest, err)
		return 0, false
	}

	// Try to decode models.GetMonthReq
	idReq := &models.GetMonthReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.Unmarshal(body, idReq); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if idReq.ID != nil {
		log.WithField("id", *idReq.ID).Debug("month id is passed")
		return *idReq.ID, true
	}

	log.Debug("try to parse year and month")

	// Try to use models.GetMonthByYearAndMonthReq
	yearAndMonthReq := &models.GetMonthByYearAndMonthReq{}
	if err := json.Unmarshal(body, yearAndMonthReq); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if yearAndMonthReq.Year == nil || yearAndMonthReq.Month == nil {
		msg := "invalid request: no id or year and month were passed"
		s.processError(r.Context(), log, w, msg, http.StatusBadRequest, nil)
		return 0, false
	}
	year := *yearAndMonthReq.Year
	month := int(*yearAndMonthReq.Month)
	log = log.WithFields(logrus.Fields{"year": year, "month": month})

	log.Debug("try to get month id")
	id, err = s.db.GetMonthID(r.Context(), year, month)
	if err != nil {
		switch err {
		case db.ErrMonthNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get month with passed year and month"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return 0, false
	}

	return id, true
}

// GET /api/days
//
// Request: models.GetDayReq or models.GetDayByDate
// Response: models.GetDayResp or models.Response
//
func (s Server) GetDay(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

	// Prepare
	dayID, ok := s.getDayID(w, r)
	if !ok {
		// 'Server.getDayID' has already called 'Server.processError'
		return
	}
	log = log.WithField("day_id", dayID)

	// Process
	log.Debug("get day from the database")
	day, err := s.db.GetDay(r.Context(), dayID)
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

func (s Server) getDayID(w http.ResponseWriter, r *http.Request) (id uint, ok bool) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.processError(r.Context(), log, w, "couldn't read body", http.StatusBadRequest, err)
		return 0, false
	}

	// Try to decode models.GetDayReq
	idReq := &models.GetDayReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.Unmarshal(body, idReq); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if idReq.ID != nil {
		log.WithField("id", idReq.ID).Debug("day id is passed")
		return *idReq.ID, true
	}

	log.Debug("try to parse year, month and day")

	// Try to use models.GetDayByDateReq
	dateReq := &models.GetDayByDateReq{}
	if err := json.Unmarshal(body, dateReq); err != nil {
		s.processError(r.Context(), log, w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if dateReq.Year == nil || dateReq.Month == nil || dateReq.Day == nil {
		s.processError(r.Context(), log, w, "invalid request: no id or year, month and day were passed",
			http.StatusBadRequest, nil)
		return 0, false
	}
	year := *dateReq.Year
	month := int(*dateReq.Month)
	day := *dateReq.Day
	log = log.WithFields(logrus.Fields{"year": year, "month": month, "day": day})

	log.Debug("try to get day id")
	id, err = s.db.GetDayIDByDate(r.Context(), year, month, day)
	if err != nil {
		switch err {
		case db.ErrDayNotExist:
			s.processError(r.Context(), log, w, err.Error(), http.StatusNotFound, nil)
		default:
			msg := "couldn't get such Day"
			s.processError(r.Context(), log, w, msg, http.StatusInternalServerError, err)
		}
		return 0, false
	}

	return id, true
}

// -------------------------------------------------
// Income
// -------------------------------------------------

// POST /api/incomes - add a new income
//
// Request: models.AddIncomeReq
// Response: models.AddIncomeResp or models.Response
//
func (s Server) AddIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// PUT /api/incomes - edit existing income
//
// Request: models.EditIncomeReq
// Response: models.Response
//
func (s Server) EditIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// DELETE /api/incomes - remove income
//
// Request: models.RemoveIncomeReq
// Response: models.Response
//
func (s Server) RemoveIncome(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// POST /api/monthly-payments
//
// Request: models.AddMonthlyPaymentReq
// Response: models.AddMonthlyPaymentResp or models.Response
//
func (s Server) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// PUT /api/monthly-payments
//
// Request: models.EditMonthlyPaymentReq
// Response: models.Response
//
func (s Server) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// DELETE /api/monthly-payments
//
// Request: models.DeleteMonthlyPaymentReq
// Response: models.Response
//
func (s Server) RemoveMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// POST /api/spends
//
// Request: models.AddSpendReq
// Response: models.AddSpendResp or models.Response
//
func (s Server) AddSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// PUT /api/spends
//
// Request: models.EditSpendReq
// Response: models.Response
//
func (s Server) EditSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// DELETE /api/spends
//
// Request: models.RemoveSpendReq
// Response: models.Response
//
func (s Server) RemoveSpend(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// GET /api/spend-types
//
// Request: -
// Response: models.GetSpendTypesResp or models.Response
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

// POST /api/spend-types
//
// Request: models.AddSpendTypeReq
// Response: models.AddSpendTypeResp or models.Response
//
func (s Server) AddSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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

// PUT /api/spend-types
//
// Request: models.EditSpendTypeReq
// Response: models.Response
//
func (s Server) EditSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// DELETE /api/spend-types
//
// Request: models.RemoveSpendTypeReq
// Response: models.Response
//
func (s Server) RemoveSpendType(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
			s.processError(r.Context(), log, w, err.Error(), http.StatusBadRequest, err)
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

// GET /api/search/spends
//
// Request: models.SearchSpendsReq
// Response: models.SearchSpendsResp
//
func (s Server) SearchSpends(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	defer r.Body.Close()

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
		args.Sort = db.SortByTitle
	case "cost":
		args.Sort = db.SortByCost
	}
	if req.Order == "desc" {
		args.Order = db.OrderByDesc
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
