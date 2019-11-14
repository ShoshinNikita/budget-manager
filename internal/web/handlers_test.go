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
	defer stopServer(requireGlobal, server)

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
				desc: "valid request with notes",
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
				}
			})
		}
	})

	t.Run("EditIncome", func(t *testing.T) {
		newTitle := func(s string) *string { return &s }
		newNotes := func(s string) *string { return &s }
		newIncome := func(i int64) *int64 { return &i }

		tests := []struct {
			desc       string
			req        models.EditIncomeReq
			statusCode int
			resp       models.Response
		}{
			{
				desc: "edit title",
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
				desc: "edit all fields",
				req: models.EditIncomeReq{
					ID:     2,
					Title:  newTitle("edited title"),
					Notes:  newNotes("updated notes"),
					Income: newIncome(123456),
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
				desc: "remove income with id '1'",
				req: models.RemoveIncomeReq{
					ID: 1,
				},
				statusCode: http.StatusOK,
				resp: models.Response{
					Success: true,
				},
			},
			{
				desc: "remove income with id '2'",
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
				desc: "remove income with id '3'",
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

func stopServer(require *require.Assertions, server *Server) {
	err := server.db.Shutdown()
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
