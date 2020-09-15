package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"
)

//nolint:gochecknoglobals
var queryDecoder = schema.NewDecoder()

//nolint:gochecknoinits
func init() {
	queryDecoder.IgnoreUnknownKeys(true)
	queryDecoder.SetAliasTag("json")
}

type RequestChecker interface {
	Check() error
}

// DecodeRequest decodes request and checks its validity. It process error if needed
func DecodeRequest(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, req RequestChecker) (ok bool) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		ProcessError(ctx, w, "couldn't parse form: "+err.Error(), http.StatusBadRequest)
		return false
	}

	var err error
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		err = decodeQueryRequest(r.Form, req)
	default:
		err = decodeJSONRequest(r.Body, req)
	}
	if err != nil {
		ProcessError(ctx, w, "couldn't decode request: "+err.Error(), http.StatusBadRequest)
		return false
	}

	if err := req.Check(); err != nil {
		ProcessError(ctx, w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func decodeJSONRequest(body io.Reader, req interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(req)
}

func decodeQueryRequest(form url.Values, req interface{}) error {
	return queryDecoder.Decode(req, form)
}
