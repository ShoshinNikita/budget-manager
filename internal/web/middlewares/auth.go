package middlewares

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

type Credentials interface {
	Get(username string) (secret string, ok bool)
}

func BasicAuthMiddleware(h http.Handler, creds Credentials, log logger.Logger) http.Handler {
	checkAuth := func(r *http.Request) bool {
		username, password, ok := r.BasicAuth()
		if !ok {
			return false
		}
		hashedPassword, ok := creds.Get(username)
		if !ok {
			return false
		}

		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx, log).WithField("ip", r.RemoteAddr)

		if !checkAuth(r) {
			log.Warn("invalid auth request")

			w.Header().Set("WWW-Authenticate", `Basic realm="Budget Manager"`)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, http.StatusText(http.StatusUnauthorized))
			return
		}

		log.Debug("successful auth request")
		h.ServeHTTP(w, r)
	})
}
