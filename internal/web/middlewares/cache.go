package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

// CachingMiddleware sets "Cache-Control" and "ETag" headers
func CachingMiddleware(h http.Handler, maxAge time.Duration, gitHash string) http.Handler {
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
