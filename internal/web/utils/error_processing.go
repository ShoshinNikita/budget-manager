package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

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
