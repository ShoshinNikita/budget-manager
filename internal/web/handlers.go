package web

import (
	"encoding/json"
	"net/http"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/db/money"
	"github.com/ShoshinNikita/budget_manager/internal/web/models"
)

const (
	errDecodeRequest  = "couldn't decode request"
	errEncodeResponse = "couldn't encode response"
)

// -------------------------------------------------
// Income
// -------------------------------------------------

// POST /api/incomes - add a new income
//
// Request: models.AddIncomeReq
// Response: models.AddIncomeResp or models.Response
//
func (s Server) AddIncome(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.AddIncomeReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		Notes:   req.Notes,
		Income:  money.FromInt(req.Income),
	}
	id, err := s.db.AddIncome(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't add new Income", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.AddIncomeResp{
		Response: models.Response{
			Success: true,
		},
		ID: id,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// PUT /api/incomes - edit existing income
//
// Request: models.EditIncomeReq
// Response: models.Response
//
func (s Server) EditIncome(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.EditIncomeReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.EditIncomeArgs{
		ID:    req.ID,
		Title: req.Title,
		Notes: req.Notes,
	}
	if req.Income != nil {
		income := money.FromInt(*req.Income)
		args.Income = &income
	}
	err := s.db.EditIncome(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't edit Income", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// DELETE /api/incomes - remove income
//
// Request: models.RemoveIncomeReq
// Response: models.Response
//
func (s Server) RemoveIncome(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.RemoveIncomeReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveIncome(req.ID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't remove Income", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Monthly Payment
// -------------------------------------------------

// POST /api/monthly-payments
//
// Request: models.AddMonthlyPaymentReq
// Response models.AddMonthlyPaymentResp or models.Response
//
func (s Server) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.AddMonthlyPaymentReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddMonthlyPaymentArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		TypeID:  req.TypeID,
		Notes:   req.Notes,
		Cost:    money.FromInt(req.Cost),
	}
	id, err := s.db.AddMonthlyPayment(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't add new Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.AddMonthlyPaymentResp{
		Response: models.Response{
			Success: true,
		},
		ID: id,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// PUT /api/monthly-payments
//
// Request: models.EditMonthlyPaymentReq
// Response models.Response
//
func (s Server) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.EditMonthlyPaymentReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.EditMonthlyPaymentArgs{
		ID:     req.ID,
		Title:  req.Title,
		Notes:  req.Notes,
		TypeID: req.TypeID,
	}
	if req.Cost != nil {
		cost := money.FromInt(*req.Cost)
		args.Cost = &cost
	}
	err := s.db.EditMonthlyPayment(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't edit Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// DELETE /api/monthly-payments
//
// Request: models.DeleteMonthlyPaymentReq
// Response models.Response
//
func (s Server) RemoveMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.RemoveMonthlyPaymentReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveMonthlyPayment(req.ID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't remove Monthly Payment", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Spends
// -------------------------------------------------

func (s Server) AddSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteSpend(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Spend Types
// -------------------------------------------------

func (s Server) AddSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) EditSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

func (s Server) DeleteSpendType(w http.ResponseWriter, r *http.Request) {
	notImplementedYet(w, r)
}

// -------------------------------------------------
// Other
// -------------------------------------------------

func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet")) //nolint:errcheck
}
