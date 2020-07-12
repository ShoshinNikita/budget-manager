package web

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

// processErrorWithPage is similar to 'processError', but it shows error page instead of returning json
func (s Server) processErrorWithPage(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter,
	respMsg string, code int, internalErr error) {

	if internalErr != nil {
		utils.LogHTTPError(log, respMsg, code, internalErr)
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
		utils.ProcessError(ctx, log, w, executeErrorMessage, http.StatusInternalServerError, err)
	}
}
