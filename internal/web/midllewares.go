package web

import (
	"fmt"
	"net/http"
	"time"
)

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
