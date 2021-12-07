package middlewares

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type Credentials interface {
	Get(username string) (secret string, ok bool)
}

func BasicAuthMiddleware(h http.Handler, creds Credentials, log logger.Logger) http.Handler {
	errUnauthorized := errors.New("unauthorized")

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
		log := reqid.FromContextToLogger(ctx, log)
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
