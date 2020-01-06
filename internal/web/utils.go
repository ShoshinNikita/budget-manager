package web

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/web/models"
)

const formatMsg = "caller: '%s', msg: '%s', error: '%s'"

func (s Server) processDBError(w http.ResponseWriter, err error) {
	var (
		code          = http.StatusInternalServerError
		msg           = err.Error()
		originalError = errors.GetOriginalError(err).Error()
	)

	errType, ok := errors.GetErrorType(err)
	if !ok {
		errType = errors.UndefinedError
		msg = errors.DefaultErrorMessage
	}

	switch errType {
	case errors.UserError:
		// Don't log user errors
		code = http.StatusBadRequest
	default:
		s.log.Errorf(formatMsg, getCallerFunc(2), msg, originalError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := models.Response{Success: false, Error: msg}
	json.NewEncoder(w).Encode(resp) // nolint:errcheck
}

// processError logs error and writes models.Response. If internalErr is nil,
// it just writes models.Response
func (s Server) processError(w http.ResponseWriter, respMsg string, code int, internalErr error) {
	if internalErr != nil {
		s.log.Errorf(formatMsg, getCallerFunc(2), respMsg, internalErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := models.Response{Success: false, Error: respMsg}
	json.NewEncoder(w).Encode(resp) // nolint:errcheck
}

const prefixForTrim = "github.com/ShoshinNikita/budget-manager/"

func getCallerFunc(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		// something like "github.com/username/project/internal/web.Service.Ping"
		funcName := details.Name()

		// trim "github.com/username/project/"
		return strings.TrimPrefix(funcName, prefixForTrim)
	}

	return ""
}
