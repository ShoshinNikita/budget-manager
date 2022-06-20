package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

// LoggingMiddleware logs HTTP requests. Logs include execution time, content length and status code
func LoggingMiddleware(h http.Handler, log logger.Logger) http.Handler {
	shouldSkip := func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/static/")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldSkip(r) {
			h.ServeHTTP(w, r)
			return
		}

		log := logger.FromContext(r.Context(), log)
		log = log.WithFields(logger.Fields{"method": r.Method, "url": r.URL.Path})

		log.Debug("start request")

		respWriter := newResponseWriter(w)
		now := time.Now()
		h.ServeHTTP(respWriter, r)
		since := time.Since(now)

		log.WithFields(logger.Fields{
			"time":           since,
			"status_code":    respWriter.statusCode,
			"content_length": respWriter.contentLength,
		}).Debug("finish request")
	})
}

// responseWriter implents 'http.ResponseWriter' interface and contains the information
// about status code and content length
type responseWriter struct {
	http.ResponseWriter

	statusCode    int
	contentLength int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.contentLength += len(data)
	return w.ResponseWriter.Write(data)
}
