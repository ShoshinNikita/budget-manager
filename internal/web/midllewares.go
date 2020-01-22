package web

import (
	"fmt"
	"net/http"
	"time"

	auth "github.com/abbot/go-http-auth"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// basicAuthMiddleware checks whether user is authorized
func (s Server) basicAuthMiddleware(h http.Handler) http.Handler {
	const realm = "Budget Manager"

	basicAuthenticator := auth.NewBasicAuthenticator(realm, func(user, _ string) string {
		if pass, ok := s.config.Credentials[user]; ok {
			return pass
		}
		return ""
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := request_id.FromContextToLogger(r.Context(), s.log)
		log = log.WithFields(logrus.Fields{"ip": r.RemoteAddr, "url": r.URL})

		if username := basicAuthenticator.CheckAuth(r); username == "" {
			// Auth has failed
			log.Warn("invalid auth request")
			basicAuthenticator.RequireAuth(w, r)
			return
		}

		log.Debug("successful auth request")
		h.ServeHTTP(w, r)
	})
}

// cacheMiddleware sets "Cache-Control" header
func cacheMiddleware(h http.Handler, maxAge time.Duration) http.Handler {
	maxAgeString := fmt.Sprintf("max-age=%d", int64(maxAge.Seconds()))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expTime := time.Now().Add(maxAge)

		w.Header().Set("Expires", expTime.Format(http.TimeFormat))
		w.Header().Set("Cache-Control", "private, "+maxAgeString)

		h.ServeHTTP(w, r)
	})
}

const requestIDHeader = "X-Request-ID"

// requestIDMeddleware generates a new request id and inserts it into the request context
func (Server) requestIDMeddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := request_id.New()
		if headerValue := r.Header.Get(requestIDHeader); headerValue != "" {
			reqID = request_id.RequestID(headerValue)
		}

		ctx := request_id.ToContext(r.Context(), reqID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all HTTP requests. It also logs execution time, content length and status code
func (s Server) loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := request_id.FromContextToLogger(r.Context(), s.log)
		log = log.WithFields(logrus.Fields{"method": r.Method, "url": r.URL.Path})

		log.Debug("start request")

		respWriter := newResponseWriter(w)
		now := time.Now()
		h.ServeHTTP(respWriter, r)
		since := time.Since(now)

		log.WithFields(logrus.Fields{
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
