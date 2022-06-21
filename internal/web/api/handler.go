package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/validator"
)

type (
	emptyReq  struct{}
	emptyResp struct{}
)

// Handler allows to use generic HandleFunc as http.Handler
type Handler[Req, Resp any] struct {
	log     logger.Logger
	handler HandleFunc[Req, Resp]
}

type HandleFunc[Req, Resp any] func(context.Context, *Req) (*Resp, error)

func NewHandler[Req, Resp any](log logger.Logger, handler HandleFunc[Req, Resp]) http.Handler {
	return &Handler[Req, Resp]{
		log:     log,
		handler: handler,
	}
}

func (h *Handler[Req, Resp]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req Req
	if !h.decodeRequest(w, r, &req) {
		return
	}

	resp, err := h.handler(ctx, &req)
	if err != nil {
		h.encodeError(ctx, w, err)
		return
	}

	h.encodeResponse(ctx, w, http.StatusOK, resp)
}

func (h *Handler[Req, Resp]) decodeRequest(w http.ResponseWriter, r *http.Request, req any) bool {
	ctx := r.Context()

	if r.Body != http.NoBody {
		// Decode requests only with non-empty body
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(req); err != nil {
			h.encodeError(ctx, w, app.NewUserError(errors.Wrap(err, "decode error")))
			return false
		}
	}

	// Validate the request even if body was empty
	if err := validator.Validate(req); err != nil {
		h.encodeError(ctx, w, app.NewUserError(err))
		return false
	}

	return true
}

func (h *Handler[Req, Resp]) encodeResponse(ctx context.Context, w http.ResponseWriter, statusCode int, resp any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(fmt.Sprintf("couldn't encode response: %s", err))
		}
	}
}

func (h *Handler[Req, Resp]) encodeError(ctx context.Context, w http.ResponseWriter, err error) {
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

		logger.FromContext(ctx, h.log).WithError(err).Error("request failed")
	}

	resp := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}
	h.encodeResponse(ctx, w, statusCode, resp)
}
