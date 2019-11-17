// +build integration

package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ShoshinNikita/go-clog/v3"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget_manager/internal/db"
	"github.com/ShoshinNikita/budget_manager/internal/web/models"
)

const (
	dbHost     = "localhost"
	dbPort     = "5432"
	dbUser     = "postgres"
	dbPassword = ""
	dbDatabase = "postgres"
)

func TestHandlers_Income(t *testing.T) {
	requireGlobal := require.New(t)
	server := initServer(requireGlobal)
	defer cleanUp(requireGlobal, server)

	t.Run("AddIncome", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.AddIncomeReq
			statusCode int
			resp       interface{}
		}{
			{
				desc: "valid request",
				req: models.AddIncomeReq{
					MonthID: 1,
					Title:   "some income",
					Income:  15000,
				},
				statusCode: http.StatusOK,
				resp: models.AddIncomeResp{
					Response: models.Response{
						Success: true,
					},
					ID: 1,
				},
			},
			{
				desc: "valid request (with notes)",
				req: models.AddIncomeReq{
					MonthID: 1,
					Title:   "some income",
					Notes:   "some notes",
					Income:  15000,
				},
				statusCode: http.StatusOK,
				resp: models.AddIncomeResp{
					Response: models.Response{
						Success: true,
					},
					ID: 2,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty title)",
				req: models.AddIncomeReq{
					MonthID: 1,
					Title:   "",
					Notes:   "some notes",
					Income:  15000,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero income)",
				req: models.AddIncomeReq{
					MonthID: 1,
					Title:   "title",
					Notes:   "some notes",
					Income:  0,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (invalid month)",
				req: models.AddIncomeReq{
					MonthID: 10,
					Title:   "title",
					Notes:   "some notes",
					Income:  15000,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("POST", "/api/incomes", body)

				// Send Request
				server.AddIncome(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				wantResp := tt.resp
				switch wantResp.(type) {
				case models.AddIncomeResp:
					resp := &models.AddIncomeResp{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				case models.Response:
					resp := &models.Response{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				default:
					require.Fail("invalid resp type")
				}
			})
		}
	})

	t.Run("EditIncome", func(t *testing.T) {
		newTitle := func(s string) *string { return &s }
		newNotes := func(s string) *string { return &s }
		newIncome := func(i float64) *float64 { return &i }

		tests := []struct {
			desc       string
			req        models.EditIncomeReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (edit title)",
				req: models.EditIncomeReq{
					ID:    1,
					Title: newTitle("edited title"),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (edit all fields)",
				req: models.EditIncomeReq{
					ID:     2,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					Income: newIncome(123456.20),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (invalid id)",
				req: models.EditIncomeReq{
					ID:     5,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					Income: newIncome(0),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero income)",
				req: models.EditIncomeReq{
					ID:     2,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					Income: newIncome(0),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("PUT", "/api/incomes", body)

				// Send Request
				server.EditIncome(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})

	t.Run("RemoveIncome", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.RemoveIncomeReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (remove Income with id '1')",
				req: models.RemoveIncomeReq{
					ID: 1,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (remove Income with id '2')",
				req: models.RemoveIncomeReq{
					ID: 2,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (remove non-existing Income)",
				req: models.RemoveIncomeReq{
					ID: 3,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("DELETE", "/api/incomes", body)

				// Send Request
				server.RemoveIncome(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})
}

func TestHandlers_MonthlyPayment(t *testing.T) {
	requireGlobal := require.New(t)
	server := initServer(requireGlobal)
	defer cleanUp(requireGlobal, server)

	t.Run("AddMonthlyPayment", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.AddMonthlyPaymentReq
			statusCode int
			resp       interface{}
		}{
			{
				desc: "valid request",
				req: models.AddMonthlyPaymentReq{
					MonthID: 1,
					Title:   "Patreon",
					Cost:    750,
				},
				statusCode: http.StatusOK,
				resp: models.AddMonthlyPaymentResp{
					Response: models.Response{
						Success: true,
					},
					ID: 1,
				},
			},
			{
				desc: "valid request (with notes and type)",
				req: models.AddMonthlyPaymentReq{
					MonthID: 1,
					Title:   "Rent",
					Notes:   "some notes",
					TypeID:  1,
					Cost:    7000,
				},
				statusCode: http.StatusOK,
				resp: models.AddMonthlyPaymentResp{
					Response: models.Response{
						Success: true,
					},
					ID: 2,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty title)",
				req: models.AddMonthlyPaymentReq{
					MonthID: 1,
					Title:   "",
					Notes:   "some notes",
					Cost:    15000,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero cost)",
				req: models.AddMonthlyPaymentReq{
					MonthID: 1,
					Title:   "title",
					Notes:   "some notes",
					Cost:    0,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (invalid month)",
				req: models.AddMonthlyPaymentReq{
					MonthID: 10,
					Title:   "title",
					Notes:   "some notes",
					Cost:    15000,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("POST", "/api/monthly-payments", body)

				// Send Request
				server.AddMonthlyPayment(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				wantResp := tt.resp
				switch wantResp.(type) {
				case models.AddMonthlyPaymentResp:
					resp := &models.AddMonthlyPaymentResp{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				case models.Response:
					resp := &models.Response{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				default:
					require.Fail("invalid resp type")
				}
			})
		}
	})

	t.Run("EditMonthlyPayment", func(t *testing.T) {
		newTitle := func(s string) *string { return &s }
		newNotes := func(s string) *string { return &s }
		newTypeID := func(u uint) *uint { return &u }
		newCost := func(i float64) *float64 { return &i }

		tests := []struct {
			desc       string
			req        models.EditMonthlyPaymentReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (edit title)",
				req: models.EditMonthlyPaymentReq{
					ID:    1,
					Title: newTitle("edited title"),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (edit all fields)",
				req: models.EditMonthlyPaymentReq{
					ID:     2,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					TypeID: newTypeID(1),
					Cost:   newCost(123456.50),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (invalid id)",
				req: models.EditMonthlyPaymentReq{
					ID:    5,
					Title: newTitle("edited title"),
					Notes: newNotes("updated notes"),
					Cost:  newCost(10.05),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero cost)",
				req: models.EditMonthlyPaymentReq{
					ID:    2,
					Title: newTitle("edited title"),
					Notes: newNotes("updated notes"),
					Cost:  newCost(0),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("PUT", "/api/monthly-payments", body)

				// Send Request
				server.EditMonthlyPayment(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})

	t.Run("RemoveMonthlyPayment", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.RemoveMonthlyPaymentReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (remove Monthly Payment with id '1')",
				req: models.RemoveMonthlyPaymentReq{
					ID: 1,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (remove Monthly Payment with id '2')",
				req: models.RemoveMonthlyPaymentReq{
					ID: 2,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (remove non-existing Monthly Payment)",
				req: models.RemoveMonthlyPaymentReq{
					ID: 3,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("DELETE", "/api/monthly-payment", body)

				// Send Request
				server.RemoveMonthlyPayment(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})
}

func TestHandlers_Spend(t *testing.T) {
	requireGlobal := require.New(t)
	server := initServer(requireGlobal)
	defer cleanUp(requireGlobal, server)

	t.Run("AddSpend", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.AddSpendReq
			statusCode int
			resp       interface{}
		}{
			{
				desc: "valid request",
				req: models.AddSpendReq{
					DayID: 1,
					Title: "Break",
					Cost:  30,
				},
				statusCode: http.StatusOK,
				resp: models.AddSpendResp{
					Response: models.Response{
						Success: true,
					},
					ID: 1,
				},
			},
			{
				desc: "valid request (with notes and type)",
				req: models.AddSpendReq{
					DayID:  10,
					Title:  "Bread",
					Notes:  "warm",
					TypeID: 1,
					Cost:   50,
				},
				statusCode: http.StatusOK,
				resp: models.AddSpendResp{
					Response: models.Response{
						Success: true,
					},
					ID: 2,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty title)",
				req: models.AddSpendReq{
					DayID: 1,
					Title: "",
					Notes: "some notes",
					Cost:  20,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero cost)",
				req: models.AddSpendReq{
					DayID: 12,
					Title: "title",
					Notes: "some notes",
					Cost:  0,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (invalid day)",
				req: models.AddSpendReq{
					DayID: 36000,
					Title: "title",
					Notes: "some notes",
					Cost:  15000,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("POST", "/api/spends", body)

				// Send Request
				server.AddSpend(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				wantResp := tt.resp
				switch wantResp.(type) {
				case models.AddSpendResp:
					resp := &models.AddSpendResp{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				case models.Response:
					resp := &models.Response{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				default:
					require.Fail("invalid resp type")
				}
			})
		}
	})

	t.Run("EditSpend", func(t *testing.T) {
		newTitle := func(s string) *string { return &s }
		newNotes := func(s string) *string { return &s }
		newTypeID := func(u uint) *uint { return &u }
		newCost := func(i float64) *float64 { return &i }

		tests := []struct {
			desc       string
			req        models.EditSpendReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (edit title)",
				req: models.EditSpendReq{
					ID:    1,
					Title: newTitle("edited title"),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (edit all fields)",
				req: models.EditSpendReq{
					ID:     2,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					TypeID: newTypeID(2),
					Cost:   newCost(0.30),
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty title)",
				req: models.EditSpendReq{
					ID:    1,
					Title: newTitle(""),
					Notes: newNotes("updated notes"),
					Cost:  newCost(10),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (invalid id)",
				req: models.EditSpendReq{
					ID:    5,
					Title: newTitle("edited title"),
					Notes: newNotes("updated notes"),
					Cost:  newCost(10),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
			{
				desc: "invalid request (zero cost)",
				req: models.EditSpendReq{
					ID:    2,
					Title: newTitle("edited title"),
					Notes: newNotes("updated notes"),
					Cost:  newCost(0),
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("PUT", "/api/spends", body)

				// Send Request
				server.EditSpend(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})

	t.Run("RemoveSpend", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.RemoveSpendReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (remove Spend with id '1')",
				req: models.RemoveSpendReq{
					ID: 1,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (remove Spend with id '2')",
				req: models.RemoveSpendReq{
					ID: 2,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (remove non-existing Spend)",
				req: models.RemoveSpendReq{
					ID: 3,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("DELETE", "/api/spends", body)

				// Send Request
				server.RemoveSpend(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})
}

func TestHandlers_SpendType(t *testing.T) {
	requireGlobal := require.New(t)
	server := initServer(requireGlobal)
	defer cleanUp(requireGlobal, server)

	t.Run("AddSpendType", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.AddSpendTypeReq
			statusCode int
			resp       interface{}
		}{
			{
				desc: "valid request",
				req: models.AddSpendTypeReq{
					Name: "first type",
				},
				statusCode: http.StatusOK,
				resp: models.AddSpendTypeResp{
					Response: models.Response{
						Success: true,
					},
					ID: 1,
				},
			},
			{
				desc: "valid request (with notes and type)",
				req: models.AddSpendTypeReq{
					Name: "second type",
				},
				statusCode: http.StatusOK,
				resp: models.AddSpendTypeResp{
					Response: models.Response{
						Success: true,
					},
					ID: 2,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty name)",
				req: models.AddSpendTypeReq{
					Name: "",
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("POST", "/api/spend-types", body)

				// Send Request
				server.AddSpendType(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				wantResp := tt.resp
				switch wantResp.(type) {
				case models.AddSpendTypeResp:
					resp := &models.AddSpendTypeResp{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				case models.Response:
					resp := &models.Response{}
					decodeResponse(require, response.Body, resp)
					response.Body.Close()

					require.Equal(wantResp, *resp)
				default:
					require.Fail("invalid resp type")
				}
			})
		}
	})

	t.Run("EditSpendType", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.EditSpendTypeReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (edit name)",
				req: models.EditSpendTypeReq{
					ID:   1,
					Name: "updated name",
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (empty name)",
				req: models.EditSpendTypeReq{
					ID:   2,
					Name: "",
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("PUT", "/api/spend-types", body)

				// Send Request
				server.EditSpendType(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})

	t.Run("RemoveSpendType", func(t *testing.T) {
		tests := []struct {
			desc       string
			req        models.RemoveSpendTypeReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "valid request (remove Spend Type with id '1')",
				req: models.RemoveSpendTypeReq{
					ID: 1,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "valid request (remove Spend Type with id '2')",
				req: models.RemoveSpendTypeReq{
					ID: 2,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			// Invalid requests
			{
				desc: "invalid request (remove non-existing Spend Type)",
				req: models.RemoveSpendTypeReq{
					ID: 3,
				},
				statusCode: http.StatusBadRequest,
				resp: models.Response{
					Error:   "bad params",
					Success: false,
				},
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.desc, func(t *testing.T) {
				// Prepare
				require := require.New(t)
				w := httptest.NewRecorder()

				// Prepare request
				body := encodeRequest(require, tt.req)
				request := httptest.NewRequest("DELETE", "/api/spend-types", body)

				// Send Request
				server.RemoveSpendType(w, request)

				// Check Response
				response := w.Result()
				require.Equal(tt.statusCode, response.StatusCode)

				resp := &models.Response{}
				decodeResponse(require, response.Body, resp)
				response.Body.Close()

				require.Equal(tt.resp, *resp)
			})
		}
	})
}

// -------------------------------------------------
// Server
// -------------------------------------------------

func initServer(require *require.Assertions) *Server {
	// Logger
	log := clog.NewDevLogger()

	// DB
	dbOpts := db.NewDBOptions{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Database: dbDatabase,
	}
	db, err := db.NewDB(dbOpts, log)
	require.Nil(err)
	err = db.DropDB()
	require.Nil(err)
	err = db.Prepare()
	require.Nil(err)

	// Server
	serverOpts := NewServerOptions{Port: ":8080"}
	server := NewServer(serverOpts, db, log)

	return server
}

func cleanUp(require *require.Assertions, server *Server) {
	err := server.db.DropDB()
	require.Nil(err)

	err = server.db.Shutdown()
	require.Nil(err)

	// There's nothing to shutdown
	// err = server.Shutdown()
	// require.Nil(err)
}

// -------------------------------------------------
// Decoding and Encoding
// -------------------------------------------------

func encodeRequest(require *require.Assertions, req interface{}) io.Reader {
	buff := bytes.NewBuffer(nil)

	err := json.NewEncoder(buff).Encode(req)
	require.Nil(err)

	return buff
}

func decodeResponse(require *require.Assertions, body io.Reader, target interface{}) {
	err := json.NewDecoder(body).Decode(target)
	require.Nil(err)
}
