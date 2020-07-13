package web

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestSearchSpends(t *testing.T) {
	t.Parallel()

	const (
		method = "GET"
		target = "/api/search/spends"
	)

	tests := []struct {
		desc string
		//
		query  url.Values
		expect func(*MockDatabase)
		//
		statusCode int
		resp       models.SearchSpendsResp
	}{
		{
			desc: "all fields are empty",
			//
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{}
				m.On("SearchSpends", mock.Anything, args).Return([]*db.Spend{}, nil)
			},
			//
			statusCode: 200,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: true},
				Spends:   []*db.Spend{},
			},
		},
		{
			desc: "pass all fields",
			//
			query: url.Values{
				"title":         {"Title"},
				"title_exactly": {"true"},
				"notes":         {"NOTES"},
				"min_cost":      {"150.55"},
				"max_cost":      {"1000.89"},
				"after":         {"2020-02-20T00:00:00Z"},
				"type_ids":      {"1", "2", "3"},
			},
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{
					Title:        "title",
					TitleExactly: true,
					Notes:        "notes",
					MinCost:      money.FromFloat(150.55),
					MaxCost:      money.FromFloat(1000.89),
					After:        time.Date(2020, 2, 20, 0, 0, 0, 0, time.UTC),
					TypeIDs:      []uint{1, 2, 3},
				}
				m.On("SearchSpends", mock.Anything, args).Return([]*db.Spend{}, nil)
			},
			//
			statusCode: 200,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: true},
				Spends:   []*db.Spend{},
			},
		},
		{
			desc: "pass without type",
			//
			query: url.Values{
				"without_type": {"true"},
				"type_ids":     {"1", "2", "3"},
			},
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{
					WithoutType: true,
				}
				m.On("SearchSpends", mock.Anything, args).Return([]*db.Spend{}, nil)
			},
			//
			statusCode: 200,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: true},
				Spends:   []*db.Spend{},
			},
		},
		{
			desc: "sort by title, desc",
			//
			query: url.Values{
				"sort":  {"title"},
				"order": {"desc"},
			},
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{
					Sort:  db.SortSpendsByTitle,
					Order: db.OrderByDesc,
				}
				m.On("SearchSpends", mock.Anything, args).Return([]*db.Spend{}, nil)
			},
			//
			statusCode: 200,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: true},
				Spends:   []*db.Spend{},
			},
		},
		{
			desc: "sort by cost",
			//
			query: url.Values{
				"sort":  {"cost"},
				"order": {"abcde"},
			},
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{
					Sort: db.SortSpendsByCost,
				}
				m.On("SearchSpends", mock.Anything, args).Return([]*db.Spend{}, nil)
			},
			//
			statusCode: 200,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: true},
				Spends:   []*db.Spend{},
			},
		},
		{
			desc: "db error",
			//
			expect: func(m *MockDatabase) {
				args := db.SearchSpendsArgs{}
				err := errors.New("internal db error")
				m.On("SearchSpends", mock.Anything, args).Return(nil, err)
			},
			//
			statusCode: 500,
			resp: models.SearchSpendsResp{
				Response: models.Response{RequestID: "request-id", Success: false, Error: "couldn't search for Spends"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			require := require.New(t)

			server, mock := prepareServer(t)

			// Prepare mock
			if tt.expect != nil {
				tt.expect(mock)
			}

			// Prepare request
			w := httptest.NewRecorder()
			//
			request := httptest.NewRequest(method, target, nil)
			request.URL.RawQuery = tt.query.Encode()
			request.Header.Set(requestIDHeader, requestID.ToString())

			// Send request
			server.server.Handler.ServeHTTP(w, request)

			// Check response
			response := w.Result()
			require.Equal(tt.statusCode, response.StatusCode)

			resp := models.SearchSpendsResp{}
			decodeResponse(require, response.Body, &resp)
			response.Body.Close()

			require.Equal(tt.resp, resp)
		})
	}
}

func prepareServer(t *testing.T) (*Server, *MockDatabase) { // nolint:unparam
	// Logger
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	// Discard log messages in tests
	log.SetOutput(ioutil.Discard)

	// DB
	dbMock := &MockDatabase{}

	// Server
	config := Config{Port: 8080, SkipAuth: true}
	server := NewServer(config, dbMock, log)
	server.Prepare()

	return server, dbMock
}
