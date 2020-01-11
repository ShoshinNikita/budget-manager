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

// parseDBError parses DB error and returns vars for passing into 'Server.processError' method
func (s Server) parseDBError(err error) (msg string, code int, originalErr error) {
	msg = err.Error()
	code = http.StatusInternalServerError
	originalErr = errors.GetOriginalError(err)

	errType, ok := errors.GetErrorType(err)
	if !ok {
		errType = errors.UndefinedError
		msg = errors.DefaultErrorMessage
	}

	if errType == errors.UserError {
		code = http.StatusBadRequest
		// 'originalErr' and 'msg' contain the same message. So, we can set 'originalErr' to nil
		// to skip logging in 'processError' method
		originalErr = nil
	}

	return msg, code, originalErr
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
