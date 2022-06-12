package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ShoshinNikita/budget-manager/v2/internal/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/utils/schema"
)

type Request interface {
	SanitizeAndCheck() error
}

// DecodeRequest decodes request and checks its validity. It process error if needed
func DecodeRequest(w http.ResponseWriter, r *http.Request, log logger.Logger, req Request) (ok bool) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		EncodeError(ctx, w, log, errors.Wrap(err, "couldn't parse form"), http.StatusBadRequest)
		return false
	}

	var err error
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		err = decodeQueryRequest(r.Form, req)
	default:
		err = decodeJSONRequest(r.Body, req)
	}
	if err != nil {
		EncodeError(ctx, w, log, errors.Wrap(err, "couldn't decode request"), http.StatusBadRequest)
		return false
	}

	if err := req.SanitizeAndCheck(); err != nil {
		EncodeError(ctx, w, log, err, http.StatusBadRequest)
		return false
	}

	return true
}

func decodeJSONRequest(body io.Reader, req interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(req)
}

func decodeQueryRequest(form url.Values, req interface{}) error {
	return schema.Decode(req, form)
}
