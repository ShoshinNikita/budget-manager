package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

// ProcessError writes 'models.Response' with passed message and status code
func ProcessError(ctx context.Context, w http.ResponseWriter, respMsg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   false,
		Error:     respMsg,
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// ProcessError logs internal error with 'LogInternalError' and calls 'ProcessError'
// with 'http.StatusInternalServerError' status code
func ProcessInternalError(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, err error) {

	LogInternalError(log, respMsg, err)

	ProcessError(ctx, w, respMsg, http.StatusInternalServerError)
}

func LogInternalError(log logrus.FieldLogger, respMsg string, internalErr error) {
	log.WithFields(logrus.Fields{"msg": respMsg, "error": internalErr}).Error("request error")
}
