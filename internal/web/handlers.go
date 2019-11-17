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
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		Notes:   req.Notes,
		Income:  money.FromFloat(req.Income),
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
	w.Header().Set("Content-Type", "application/json")
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
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
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
		income := money.FromFloat(*req.Income)
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
	w.Header().Set("Content-Type", "application/json")
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
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
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
	w.Header().Set("Content-Type", "application/json")
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
// Response: models.AddMonthlyPaymentResp or models.Response
//
func (s Server) AddMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.AddMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddMonthlyPaymentArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		TypeID:  req.TypeID,
		Notes:   req.Notes,
		Cost:    money.FromFloat(req.Cost),
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
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// PUT /api/monthly-payments
//
// Request: models.EditMonthlyPaymentReq
// Response: models.Response
//
func (s Server) EditMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.EditMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
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
		cost := money.FromFloat(*req.Cost)
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
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// DELETE /api/monthly-payments
//
// Request: models.DeleteMonthlyPaymentReq
// Response: models.Response
//
func (s Server) RemoveMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.RemoveMonthlyPaymentReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
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
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
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
	defer r.Body.Close()

	// Decode
	req := &models.AddSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddSpendArgs{
		DayID:  req.DayID,
		Title:  req.Title,
		TypeID: req.TypeID,
		Notes:  req.Notes,
		Cost:   money.FromFloat(req.Cost),
	}
	id, err := s.db.AddSpend(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't add new Spend", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.AddSpendResp{
		Response: models.Response{
			Success: true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// PUT /api/spends
//
// Request: models.EditSpendReq
// Response: models.Response
//
func (s Server) EditSpend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.EditSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

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
	err := s.db.EditSpend(args)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't edit Spend", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// DELETE /api/spends
//
// Request: models.RemoveSpendReq
// Response: models.Response
//
func (s Server) RemoveSpend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.RemoveSpendReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveSpend(req.ID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't remove Spend", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Spend Types
// -------------------------------------------------

// POST /api/spend-types
//
// Request: models.AddSpendTypeReq
// Response: models.AddSpendTypeResp or models.Response
//
func (s Server) AddSpendType(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.AddSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	id, err := s.db.AddSpendType(req.Name)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't add new Spend Type", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.AddSpendTypeResp{
		Response: models.Response{
			Success: true,
		},
		ID: id,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// PUT /api/spend-types
//
// Request: models.EditSpendTypeReq
// Response: models.Response
//
func (s Server) EditSpendType(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.EditSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.EditSpendType(req.ID, req.Name)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't edit Spend Type", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// DELETE /api/spend-types
//
// Request: models.RemoveSpendTypeReq
// Response: models.Response
//
func (s Server) RemoveSpendType(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Decode
	req := &models.RemoveSpendTypeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveSpendType(req.ID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad params", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't remove Spend Type", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.Response{
		Success: true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Other
// -------------------------------------------------

func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet")) //nolint:errcheck
}
