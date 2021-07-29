package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/schema"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

type Path string

const (
	IncomesPath         Path = "/api/incomes"
	MonthlyPaymentsPath Path = "/api/monthly-payments"
	SpendsPath          Path = "/api/spends"
	SpendTypesPath      Path = "/api/spend-types"
	MonthsPath          Path = "/api/months/id"
)

type Method string

const (
	GET    Method = http.MethodGet
	HEAD   Method = http.MethodHead
	POST   Method = http.MethodPost
	PUT    Method = http.MethodPut
	DELETE Method = http.MethodDelete
)

type Request struct {
	Method  Method
	Path    Path
	Request interface{}

	StatusCode int
	Err        string
}

func (t Request) Send(require *require.Assertions, host string, resp interface{}) {
	r := t.sendRequest(require, http.DefaultClient, host)
	defer r.Body.Close()

	t.checkResponse(require, r, resp)
}

func (t Request) sendRequest(require *require.Assertions, client *http.Client, host string) *http.Response {
	u := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   string(t.Path),
	}

	var body io.Reader
	if t.Request != nil {
		//nolint:exhaustive
		switch t.Method {
		case GET, HEAD:
			query := url.Values{}

			err := newQueryEncoder().Encode(t.Request, query)
			require.NoError(err, "couldn't prepare query")

			u.RawQuery = query.Encode()

		default:
			buf := &bytes.Buffer{}

			err := json.NewEncoder(buf).Encode(t.Request)
			require.NoError(err, "couldn't prepare body")

			body = buf
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), string(t.Method), u.String(), body)
	require.NoError(err)

	resp, err := client.Do(req)
	require.NoError(err, "request failed")

	return resp
}

func newQueryEncoder() *schema.Encoder {
	enc := schema.NewEncoder()
	enc.SetAliasTag("json")
	return enc
}

func (t Request) checkResponse(require *require.Assertions, r *http.Response, customResp interface{}) {
	body, err := ioutil.ReadAll(r.Body)
	require.NoError(err, "couldn't read body")

	var basicResp models.Response

	err = json.Unmarshal(body, &basicResp)
	require.NoError(err, "couldn't decode basic response")
	require.Equal(t.Err, basicResp.Error)
	require.Equal(t.Err == "", basicResp.Success)
	require.Equal(t.StatusCode, r.StatusCode)

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
