package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils/schema"
)

type Path string

const (
	IncomesPath         Path = "/api/incomes"
	MonthlyPaymentsPath Path = "/api/monthly-payments"
	SpendsPath          Path = "/api/spends"
	SpendTypesPath      Path = "/api/spend-types"
	SearchSpendsPath    Path = "/api/search/spends"
	MonthsPath          Path = "/api/months/id"
	DaysPath            Path = "/api/days/id"
)

type Method string

const (
	GET    Method = http.MethodGet
	HEAD   Method = http.MethodHead
	POST   Method = http.MethodPost
	PUT    Method = http.MethodPut
	DELETE Method = http.MethodDelete
)

type RequestOK struct {
	Method  Method
	Path    Path
	Request interface{}
}

func (r RequestOK) Send(t *testing.T, host string, resp interface{}) {
	Request{r.Method, r.Path, r.Request, http.StatusOK, ""}.Send(t, host, resp)
}

type RequestCreated struct {
	Method  Method
	Path    Path
	Request interface{}
}

func (r RequestCreated) Send(t *testing.T, host string, resp interface{}) {
	Request{r.Method, r.Path, r.Request, http.StatusCreated, ""}.Send(t, host, resp)
}

type Request struct {
	Method  Method
	Path    Path
	Request interface{}

	StatusCode int
	Err        string
}

func (r Request) Send(t *testing.T, host string, resp interface{}) {
	rawResp := r.send(t, http.DefaultClient, host)
	defer rawResp.Body.Close()

	r.checkResponse(t, rawResp, resp)
}

func (r Request) send(t *testing.T, client *http.Client, host string) *http.Response {
	require := require.New(t)

	u := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   string(r.Path),
	}

	var body io.Reader
	if r.Request != nil {
		//nolint:exhaustive
		switch r.Method {
		case GET, HEAD:
			query := url.Values{}

			err := schema.Encode(r.Request, query)
			require.NoError(err, "couldn't prepare query")

			u.RawQuery = query.Encode()

		default:
			buf := &bytes.Buffer{}

			err := json.NewEncoder(buf).Encode(r.Request)
			require.NoError(err, "couldn't prepare body")

			body = buf
		}
	}

	req, cancel := newRequest(t, r.Method, u.String(), body)
	defer cancel()

	resp, err := client.Do(req)
	require.NoError(err, "request failed")

	return resp
}

func (r Request) checkResponse(t *testing.T, resp *http.Response, customResp interface{}) {
	require := require.New(t)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(err, "couldn't read body")

	var basicResp models.Response

	err = json.Unmarshal(body, &basicResp)
	require.NoError(err, "couldn't decode basic response")
	require.Equal(r.Err, basicResp.Error)
	require.Equal(r.Err == "", basicResp.Success)
	require.Equal(r.StatusCode, resp.StatusCode)

	if customResp != nil {
		err = json.Unmarshal(body, customResp)
		require.NoErrorf(err, "couldn't decode passed response of type %T", customResp)

		resetRequestID(customResp)
	}
}

func resetRequestID(resp interface{}) {
	value := reflect.ValueOf(resp)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}

	reqID := value.FieldByName("RequestID")
	if !reqID.IsValid() {
		return
	}
	if reqID.Kind() != reflect.String {
		return
	}

	reqID.SetString("")
}
