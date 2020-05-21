package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/models"
)

// processError logs error and writes models.Response. If internalErr is nil,
// it just writes models.Response
func (s Server) processError(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		s.logError(log, respMsg, code, internalErr)
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

// processErrorWithPage is similar to 'processError', but it shows error page instead of returning json
func (s Server) processErrorWithPage(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		s.logError(log, respMsg, code, internalErr)
	}

	data := struct {
		Code      int
		RequestID request_id.RequestID
		Message   string
	}{
		Code:      code,
		RequestID: request_id.FromContext(ctx),
		Message:   respMsg,
	}
	if err := s.tplStore.Execute(ctx, errorPageTemplatePath, w, data); err != nil {
		s.processError(ctx, log, w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}

func (Server) logError(log logrus.FieldLogger, respMsg string, code int, internalErr error) {
	log = log.WithFields(logrus.Fields{"msg": respMsg, "code": code, "error": internalErr})
	log.Error("request error")
}