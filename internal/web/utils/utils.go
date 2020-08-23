package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
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
		ProcessError(ctx, log, w, "couldn't parse form: "+err.Error(), http.StatusBadRequest, nil)
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
		ProcessError(ctx, log, w, "couldn't decode request: "+err.Error(), http.StatusBadRequest, nil)
		return false
	}

	if err := req.Check(); err != nil {
		ProcessError(ctx, log, w, err.Error(), http.StatusBadRequest, nil)
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

// EncodeResponse encodes response. It process error if needed
func EncodeResponse(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, resp interface{}) (ok bool) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ProcessError(r.Context(), log, w, "couldn't encode response", http.StatusInternalServerError, err)
		return false
	}
	return true
}

// ProcessError logs error and writes models.Response. If internalErr is nil,
// it just writes models.Response
//
//nolint:gofumpt
func ProcessError(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		LogInternalError(log, respMsg, internalErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := models.Response{
		RequestID: request_id.FromContext(ctx).ToString(),
		Success:   false,
		Error:     respMsg,
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func LogInternalError(log logrus.FieldLogger, respMsg string, internalErr error) {
	log.WithFields(logrus.Fields{"msg": respMsg, "error": internalErr}).Error("request error")
}
