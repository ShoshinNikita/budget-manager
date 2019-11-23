package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/db/money"
	"github.com/ShoshinNikita/budget_manager/internal/web/models"
)

const (
	errDecodeRequest  = "couldn't decode request"
	errEncodeResponse = "couldn't encode response"
)

// GET /api/months
//
// Request: models.GetMonthReq or models.GetMonthByYearAndMonthReq
// Response: models.GetMonthResp or models.Response
//
func (s Server) GetMonth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Prepare
	var monthID uint

	body := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, body)

	// Try to decode models.GetMonthReq
	req := &models.GetMonthReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.NewDecoder(tee).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	if req.ID != nil {
		monthID = *req.ID
	} else {
		// Try to use models.GetMonthByYearAndMonthReq
		req := &models.GetMonthByYearAndMonthReq{}
		if err := jsonNewDecoder(body).Decode(req); err != nil {
			s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
			return
		}
		if req.Year == nil || req.Month == nil {
			s.processError(w, "invalid request: no id or year and month were passed", http.StatusBadRequest, nil)
			return
		}

		id, err := s.db.GetMonthID(*req.Year, int(*req.Month))
		if err != nil {
			switch {
			case db.IsBadRequestError(err):
				s.processError(w, "such Month doesn't exist", http.StatusBadRequest, err)
			default:
				s.processError(w, "can't select Month with passed data", http.StatusInternalServerError, err)
			}
			return
		}

		monthID = id
	}

	// Process
	month, err := s.db.GetMonth(monthID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "Month with passed id doesn't exist", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't select Month", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.GetMonthResp{
		Response: models.Response{
			Success: true,
		},
		Month: *month,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// GET /api/days
//
// Request: models.GetDayReq or models.GetDayByDate
// Response: models.GetDayResp or models.Response
//
func (s Server) GetDay(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Prepare
	var dayID uint

	body := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, body)

	// Try to decode models.GetDayReq
	req := &models.GetDayReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.NewDecoder(tee).Decode(req); err != nil {
		s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}
	if req.ID != nil {
		dayID = *req.ID
	} else {
		// Try to use models.GetDayByDateReq
		req := &models.GetDayByDateReq{}
		if err := jsonNewDecoder(body).Decode(req); err != nil {
			s.processError(w, errDecodeRequest, http.StatusBadRequest, err)
			return
		}
		if req.Year == nil || req.Month == nil || req.Day == nil {
			s.processError(w, "invalid request: no id or year, month and day were passed", http.StatusBadRequest, nil)
			return
		}

		id, err := s.db.GetDayIDByDate(*req.Year, int(*req.Month), *req.Day)
		if err != nil {
			switch {
			case db.IsBadRequestError(err):
				s.processError(w, "such Day doesn't exist", http.StatusBadRequest, err)
			default:
				s.processError(w, "can't select Day with passed data", http.StatusInternalServerError, err)
			}
			return
		}

		dayID = id
	}

	// Process
	day, err := s.db.GetDay(dayID)
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "Day with passed id doesn't exist", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't add select Day", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.GetDayResp{
		Response: models.Response{
			Success: true,
		},
		Day: *day,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
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

// GET /api/spend-types
//
// Request: -
// Response: models.GetSpendTypesResp or models.Response
//
func (s Server) GetSpendTypes(w http.ResponseWriter, r *http.Request) {
	// Process
	types, err := s.db.GetSpendTypes()
	if err != nil {
		switch {
		case db.IsBadRequestError(err):
			s.processError(w, "bad request", http.StatusBadRequest, err)
		default:
			s.processError(w, "can't get all Spend Types", http.StatusInternalServerError, err)
		}
		return
	}

	resp := models.GetSpendTypesResp{
		Response: models.Response{
			Success: true,
		},
		SpendTypes: types,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

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

//nolint:unused,deadcode,errcheck
func notImplementedYet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented yet"))
}

// -------------------------------------------------
// Helpers
// -------------------------------------------------

// jsonNewDecoder is a wrapper for json.NewDecoder function.
// It creates a new json.Decoder and calls json.Decoder.DisallowUnknownFields method
func jsonNewDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	return d
}
