package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/api/models"
)

type responseEncoder struct {
	resp         models.Response
	statusCode   int
	success      bool
	respErrorMsg string
}

type EncodeOption func(*responseEncoder)

// EncodeResponse can be used to encode a custom response
func EncodeResponse(resp models.Response) EncodeOption {
	return func(enc *responseEncoder) {
		enc.resp = resp
	}
}

// EncodeStatusCode can be used to write a custom status code
func EncodeStatusCode(statusCode int) EncodeOption {
	return func(enc *responseEncoder) {
		enc.statusCode = statusCode
	}
}

// Encode is a helper function to encode API responses. It writes http.StatusOK and
// encodes a base response by default. The fields of the base response are automatically filled
// with values for a "successful" response. Use encode options or other Encode... functions
// to change this behavior
func Encode(ctx context.Context, w http.ResponseWriter, log logger.Logger, options ...EncodeOption) (ok bool) {
	enc := &responseEncoder{
		resp:       &models.BaseResponse{},
		statusCode: http.StatusOK,
		success:    true,
	}
	for _, opt := range options {
		opt(enc)
	}

	enc.resp.SetBaseResponse(models.BaseResponse{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   enc.success,
		Error:     enc.respErrorMsg,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(enc.statusCode)
	if err := json.NewEncoder(w).Encode(enc.resp); err != nil {
		LogInternalError(log, "couldn't encode response", err)
		return false
	}
	return true
}

// EncodeError is a wrapper for Encode that encodes error response
func EncodeError(
	ctx context.Context, w http.ResponseWriter, log logger.Logger,
	err error, statusCode int, options ...EncodeOption,
) bool {

	options = append(options, func(enc *responseEncoder) {
		enc.statusCode = statusCode
		enc.success = false
		enc.respErrorMsg = err.Error()
	})
	return Encode(ctx, w, log, options...)
}

// EncodeInternalError is a wrapper for Encode that logs and encodes an internal error
func EncodeInternalError(
	ctx context.Context, w http.ResponseWriter, log logger.Logger,
	respMsg string, err error, options ...EncodeOption,
) bool {

	LogInternalError(log, respMsg, err)

	options = append(options, func(enc *responseEncoder) {
		enc.statusCode = http.StatusInternalServerError
		enc.success = false
		enc.respErrorMsg = respMsg
	})
	return Encode(ctx, w, log, options...)
}
