package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

// basicAuthMiddleware checks whether user is authorized
func (s Server) basicAuthMiddleware(h http.Handler) http.Handler {
	errUnauthorized := errors.New("unauthorized")

	checkAuth := func(r *http.Request) bool {
		username, password, ok := r.BasicAuth()
		if !ok {
			return false
		}
		hashedPassword, ok := s.config.Credentials[username]
		if !ok {
			return false
		}

		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := reqid.FromContextToLogger(ctx, s.log)
		log = log.WithFields(logger.Fields{"ip": r.RemoteAddr})

		if !checkAuth(r) {
			log.Warn("invalid auth request")

			w.Header().Set("WWW-Authenticate", `Basic realm="Budget Manager"`)
			utils.EncodeError(ctx, w, log, errUnauthorized, http.StatusUnauthorized)
			return
		}

		log.Debug("successful auth request")
		h.ServeHTTP(w, r)
	})
}

// cacheMiddleware sets "Cache-Control" and "ETag" headers
func cacheMiddleware(h http.Handler, maxAge time.Duration, gitHash string) http.Handler {
	cacheControl := fmt.Sprintf("private, max-age=%d", int64(maxAge.Seconds()))
	etag := fmt.Sprintf(`"%s"`, gitHash)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expTime := time.Now().Add(maxAge)

		w.Header().Set("Expires", expTime.Format(http.TimeFormat))
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("ETag", etag)

		h.ServeHTTP(w, r)
	})
}

const requestIDHeader = "X-Request-ID"

// requestIDMeddleware generates a new request id and inserts it into the request context
func (Server) requestIDMeddleware(h http.Handler) http.Handler {
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

// loggingMiddleware logs HTTP requests. Logs include execution time, content length and status code
func (s Server) loggingMiddleware(h http.Handler) http.Handler {
	shouldSkip := func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/static/")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldSkip(r) {
			h.ServeHTTP(w, r)
			return
		}

		log := reqid.FromContextToLogger(r.Context(), s.log)
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
