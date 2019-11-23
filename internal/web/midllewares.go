package web

import (
	"fmt"
	"net/http"
	"time"

	auth "github.com/abbot/go-http-auth"
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
		if username := basicAuthenticator.CheckAuth(r); username == "" {
			// Auth has failed
			s.log.Errorf("invalid auth request from '%s' to '%s'", r.RemoteAddr, r.URL)
			basicAuthenticator.RequireAuth(w, r)
			return
		}

		s.log.Debugf("successful auth from '%s' to '%s'", r.RemoteAddr, r.URL)
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
