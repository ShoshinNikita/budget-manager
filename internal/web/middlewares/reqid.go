package middlewares

import (
	"net/http"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/reqid"
)

const requestIDHeader = "X-Request-ID"

// requestIDMeddleware generates a new request id and inserts it into the request context
func RequestIDMeddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := reqid.New()
		if headerValue := r.Header.Get(requestIDHeader); headerValue != "" {
			reqID = reqid.RequestID(headerValue)
		}

		ctx := reqid.ToContext(r.Context(), reqID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}
