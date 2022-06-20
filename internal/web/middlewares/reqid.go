package middlewares

import (
	"net/http"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/reqid"
)

const requestIDHeader = "X-Request-ID"

// requestIDMeddleware generates a new request id and inserts it into the request context
func RequestIDMeddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := reqid.New()
		if headerValue := r.Header.Get(requestIDHeader); headerValue != "" {
			requestID = reqid.RequestID(headerValue)
		}

		r = r.WithContext(requestID.ToContext(r.Context()))

		w.Header().Set(requestIDHeader, string(requestID))
		h.ServeHTTP(w, r)
	})
}
