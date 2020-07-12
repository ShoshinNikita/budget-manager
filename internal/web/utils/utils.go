package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/models"
)

type RequestChecker interface {
	Check() error
}

// DecodeRequest decodes request and checks its validity. It process error if needed
func DecodeRequest(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, req RequestChecker) (ok bool) {
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		ProcessError(ctx, log, w, "couldn't decode request", http.StatusBadRequest, err)
		return false
	}

	if err := req.Check(); err != nil {
		ProcessError(ctx, log, w, err.Error(), http.StatusBadRequest, nil)
		return false
	}

	return true
}

// EncodeResponse encodes response. It process error if needed
//
// nolint:unparam
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
func ProcessError(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		LogHTTPError(log, respMsg, code, internalErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := models.Response{
		RequestID: request_id.FromContext(ctx).ToString(),
		Success:   false,
		Error:     respMsg,
	}
	json.NewEncoder(w).Encode(resp) // nolint:errcheck
}

func LogHTTPError(log logrus.FieldLogger, respMsg string, code int, internalErr error) {
	log = log.WithFields(logrus.Fields{"msg": respMsg, "code": code, "error": internalErr})
	log.Error("request error")
}
