package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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
	year, month, _ := time.Now().Date()

	url := fmt.Sprintf("/overview/%d/%d", year, month)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// GET /api/months
//
// Request: models.GetMonthReq or models.GetMonthByYearAndMonthReq
// Response: models.GetMonthResp or models.Response
//
func (s Server) GetMonth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Prepare
	var monthID uint
	monthID, ok := s.getMonthID(w, r)
	if !ok {
		// 'Server.getMonthID' has already called 'Server.processError'
		return
	}

	// Process
	month, err := s.db.GetMonth(r.Context(), monthID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

func (s Server) getMonthID(w http.ResponseWriter, r *http.Request) (id uint, ok bool) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.processError(r.Context(), w, "couldn't read body", http.StatusBadRequest, err)
		return 0, false
	}

	// Try to decode models.GetMonthReq
	idReq := &models.GetMonthReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.Unmarshal(body, idReq); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if idReq.ID != nil {
		return *idReq.ID, true
	}

	// Try to use models.GetMonthByYearAndMonthReq
	yearAndMonthReq := &models.GetMonthByYearAndMonthReq{}
	if err := json.Unmarshal(body, yearAndMonthReq); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if yearAndMonthReq.Year == nil || yearAndMonthReq.Month == nil {
		s.processError(r.Context(), w, "invalid request: no id or year and month were passed", http.StatusBadRequest, nil)
		return 0, false
	}

	id, err = s.db.GetMonthID(
		r.Context(), *yearAndMonthReq.Year, int(*yearAndMonthReq.Month),
	)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
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
	defer r.Body.Close()

	// Prepare
	dayID, ok := s.getDayID(w, r)
	if !ok {
		// 'Server.getDayID' has already called 'Server.processError'
		return
	}

	// Process
	day, err := s.db.GetDay(r.Context(), dayID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

func (s Server) getDayID(w http.ResponseWriter, r *http.Request) (id uint, ok bool) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.processError(r.Context(), w, "couldn't read body", http.StatusBadRequest, err)
		return 0, false
	}

	// Try to decode models.GetDayReq
	idReqreq := &models.GetDayReq{}
	// We have to use json.NewDecoder because there are several types of request
	if err := json.Unmarshal(body, idReqreq); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if idReqreq.ID != nil {
		return *idReqreq.ID, true
	}

	// Try to use models.GetDayByDateReq
	dateReq := &models.GetDayByDateReq{}
	if err := json.Unmarshal(body, dateReq); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return 0, false
	}
	if dateReq.Year == nil || dateReq.Month == nil || dateReq.Day == nil {
		s.processError(r.Context(), w, "invalid request: no id or year, month and day were passed",
			http.StatusBadRequest, nil)
		return 0, false
	}

	id, err = s.db.GetDayIDByDate(
		r.Context(), *dateReq.Year, int(*dateReq.Month), *dateReq.Day,
	)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
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
	defer r.Body.Close()

	// Decode
	req := &models.AddIncomeReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.AddIncomeArgs{
		MonthID: req.MonthID,
		Title:   req.Title,
		Notes:   req.Notes,
		Income:  money.FromFloat(req.Income),
	}
	id, err := s.db.AddIncome(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
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
	err := s.db.EditIncome(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveIncome(r.Context(), req.ID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
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
	id, err := s.db.AddMonthlyPayment(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
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
	err := s.db.EditMonthlyPayment(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveMonthlyPayment(r.Context(), req.ID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
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
	id, err := s.db.AddSpend(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
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
	err := s.db.EditSpend(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveSpend(r.Context(), req.ID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
	types, err := s.db.GetSpendTypes(r.Context())
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	id, err := s.db.AddSpendType(r.Context(), req.Name)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.EditSpendType(r.Context(), req.ID, req.Name)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	err := s.db.RemoveSpendType(r.Context(), req.ID)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

	resp := models.Response{
		RequestID: request_id.FromContext(r.Context()).ToString(),
		Success:   true,
	}

	// Encode
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
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
	defer r.Body.Close()

	// Decode
	req := &models.SearchSpendsReq{}
	if err := jsonNewDecoder(r.Body).Decode(req); err != nil {
		s.processError(r.Context(), w, errDecodeRequest, http.StatusBadRequest, err)
		return
	}

	// Process
	args := db.SearchSpendsArgs{
		Title:        strings.ToLower(req.Title),
		Notes:        strings.ToLower(req.Notes),
		TitleExactly: req.TitleExactly,
		NotesExactly: req.NotesExactly,
		After:        req.After,
		Before:       req.Before,
		MinCost:      money.FromFloat(req.MinCost),
		MaxCost:      money.FromFloat(req.MaxCost),
		TypeIDs:      req.TypeIDs,
	}
	spends, err := s.db.SearchSpends(r.Context(), args)
	if err != nil {
		msg, code, err := s.parseDBError(err)
		s.processError(r.Context(), w, msg, code, err)
		return
	}

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
		s.processError(r.Context(), w, errEncodeResponse, http.StatusInternalServerError, err)
	}
}

// -------------------------------------------------
// Helpers
// -------------------------------------------------

//nolint:unused,deadcode,errcheck
func notImplementedYet(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("not implemented yet"))
}

// jsonNewDecoder is a wrapper for json.NewDecoder function.
// It creates a new json.Decoder and calls json.Decoder.DisallowUnknownFields method
func jsonNewDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	return d
}
