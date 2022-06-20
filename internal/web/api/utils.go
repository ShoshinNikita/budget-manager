package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

func (api API) decodeRequest(w http.ResponseWriter, r *http.Request, req any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		api.encodeError(r.Context(), w, app.NewUserError(err))
		return false
	}
	return true
}

func (api API) encodeResponse(ctx context.Context, w http.ResponseWriter, statusCode int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(fmt.Sprintf("couldn't encode response: %s", err))
		}
	}
}

func (api API) encodeError(ctx context.Context, w http.ResponseWriter, err error) {
	var (
		statusCode int
		msg        string
	)
	if userError := app.AsUserError(err); userError != nil {
		statusCode = http.StatusBadRequest
		msg = userError.Err.Error()

	} else if notFoundError := app.AsNotFound(err); notFoundError != nil {
		statusCode = http.StatusNotFound
		msg = notFoundError.Error()

	} else if alreadyExistError := app.AsAlreadyExist(err); alreadyExistError != nil {
		statusCode = http.StatusConflict
		msg = alreadyExistError.Error()

	} else {
		statusCode = http.StatusInternalServerError
		msg = "internal error"

		logger.FromContext(ctx, api.log).WithError(err).Error("request failed")
	}

	resp := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}
	api.encodeResponse(ctx, w, statusCode, resp)
}
