package middlewares

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/totp"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type Credentials interface {
	Get(username string) (secret string, ok bool)
}

func TOTPAuthMiddleware(h http.Handler, creds Credentials, log logger.Logger) http.Handler {
	return authMiddleware{
		next:  h,
		log:   log,
		creds: creds,
		checkAuth: func(d basicAuthData) bool {
			secret := d.secretFromCreds
			reqPass := d.secretFromRequest
			return totp.Generate(secret).Equal(reqPass)
		},
	}
}

func BasicAuthMiddleware(h http.Handler, creds Credentials, log logger.Logger) http.Handler {
	return authMiddleware{
		next:  h,
		log:   log,
		creds: creds,
		checkAuth: func(d basicAuthData) bool {
			hashedPass := []byte(d.secretFromCreds)
			reqPass := []byte(d.secretFromRequest)
			return bcrypt.CompareHashAndPassword(hashedPass, reqPass) == nil
		},
	}
}

const (
	sessionCookieName = "session"
	sessionTime       = time.Hour
)

type authMiddleware struct {
	next      http.Handler
	log       logger.Logger
	creds     Credentials
	checkAuth func(basicAuthData) bool
}

func (m authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, m.log).WithFields(logger.Fields{"ip": r.RemoteAddr})

	authData, ok := m.getBasicAuthData(r)
	if !ok {
		m.writeUnauthorizedError(ctx, w, log)
		return
	}

	switch cookie, err := r.Cookie(sessionCookieName); {
	case err != nil:
		// TODO: rate limit the number of auth requests?

		if !m.checkAuth(authData) {
			m.writeUnauthorizedError(ctx, w, log)
			return
		}

		log.Info("successful auth request")

		if err := m.setSessionCookie(w, authData); err != nil {
			utils.LogInternalError(log, "auth failed", err)
			return
		}

	case !m.checkSessionCookie(cookie.Value, authData):
		m.removeSessionCookie(w)
		m.writeUnauthorizedError(ctx, w, log)
		return
	}

	m.next.ServeHTTP(w, r)
}

type basicAuthData struct {
	secretFromRequest string
	secretFromCreds   string
}

func (m authMiddleware) getBasicAuthData(r *http.Request) (basicAuthData, bool) {
	username, reqSecret, ok := r.BasicAuth()
	if !ok {
		return basicAuthData{}, false
	}
	credSecret, ok := m.creds.Get(username)
	if !ok {
		return basicAuthData{}, false
	}

	return basicAuthData{
		secretFromRequest: reqSecret,
		secretFromCreds:   credSecret,
	}, true
}

func (authMiddleware) checkSessionCookie(cookieValue string, d basicAuthData) bool {
	return bcrypt.CompareHashAndPassword([]byte(cookieValue), []byte(d.secretFromCreds)) == nil
}

func (authMiddleware) setSessionCookie(w http.ResponseWriter, d basicAuthData) error {
	cookieValue, err := bcrypt.GenerateFromPassword([]byte(d.secretFromCreds), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "couldn't generate cookie value")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    string(cookieValue),
		Path:     "/",
		MaxAge:   int(sessionTime.Seconds()),
		HttpOnly: true,
	})
	return nil
}

func (authMiddleware) removeSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func (authMiddleware) writeUnauthorizedError(ctx context.Context, w http.ResponseWriter, log logger.Logger) {
	log.Warn("invalid auth request")

	w.Header().Set("WWW-Authenticate", `Basic realm="Budget Manager"`)
	utils.EncodeError(ctx, w, log, errors.New("unauthorized"), http.StatusUnauthorized)
}
