package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/web"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/totp"
)

func TestBasicAuth(t *testing.T) {
	t.Parallel()

	RunTest(
		t,
		authTest{
			getWrongUser: func() string { return "admin" },
			getWrongPass: func() string { return "12345678" },
			//
			getCorrectUser: func() string { return "user" },
			getCorrectPass: func() string { return "qwerty" },
			//
			getSecondUser: func() string { return "test" },
		},
		func(env *TestEnv) {
			env.Cfg.Server.Auth.Disable = false
			env.Cfg.Server.Auth.Type = "basic"
			env.Cfg.Server.Auth.BasicAuthCreds = web.Credentials{
				"user": "$2y$05$wK5Ad.qdY.ZLPsfEv3rc/.uO.8SkbD6r2ptiuZefMUOX0wgGK/1rC", // user:qwerty
				"test": "$2y$05$LyfK2t0YNEvW9FzMUglIr.P9p8uhO9Wtd5aeEzhzg2BjprsRoTJz2", // test:test
			}
		})
}

func TestTOTPAuth(t *testing.T) {
	t.Parallel()

	const secret = "1b5c0708365f379a63a2be1dcca69ae0dbe026db" //nolint:gosec

	RunTest(
		t,
		authTest{
			getWrongUser: func() string { return "admin" },
			getWrongPass: func() string { return "12345678" },
			//
			getCorrectUser: func() string { return "user" },
			getCorrectPass: func() string { return string(totp.Generate(secret)) },
			//
			getSecondUser: func() string { return "test" },
		},
		func(env *TestEnv) {
			env.Cfg.Server.Auth.Disable = false
			env.Cfg.Server.Auth.Type = "totp"
			env.Cfg.Server.Auth.TOTPAuthSecrets = web.Credentials{
				"user": secret,
				"test": "3809eba2c5d3a205c6a2679cde8db0b6bafeb6b2",
			}
		})
}

type authTest struct {
	getWrongUser func() string
	getWrongPass func() string

	getCorrectUser func() string
	getCorrectPass func() string
	getSecondUser  func() string
}

func (at authTest) Test(t *testing.T, host string) {
	url := fmt.Sprintf("http://%s%s", host, SearchSpendsPath)

	jar, _ := cookiejar.New(nil)

	tests := []struct {
		name      string
		getUserFn func() string
		getPassFn func() string
		//
		wantAuthorized      bool
		wantTooManyRequests bool
	}{
		{name: "no auth"},
		{name: "wrong username and password", getUserFn: at.getWrongUser, getPassFn: at.getWrongPass},
		{name: "wrong username", getUserFn: at.getWrongUser, getPassFn: at.getCorrectPass},
		{name: "wrong password", getUserFn: at.getCorrectUser, getPassFn: at.getWrongPass},
		{name: "correct credentials", getUserFn: at.getCorrectUser, getPassFn: at.getCorrectPass, wantAuthorized: true},
		// Password doesn't matter anymore. Repeat two times just in case
		{name: "auth by cookie", getUserFn: at.getCorrectUser, wantAuthorized: true},
		{name: "auth by cookie", getUserFn: at.getCorrectUser, wantAuthorized: true},
		// Cookie shouldn't work for another user, and it has to be removed
		{name: "try to auth by another user", getUserFn: at.getSecondUser},
		// Rate limiter
		{name: "rate limit", getUserFn: at.getCorrectUser},                                   // 2 tokens left
		{name: "rate limit", getUserFn: at.getCorrectUser},                                   // 1 token left
		{name: "rate limit", getUserFn: at.getCorrectUser},                                   // 0 tokens left
		{name: "too many requests", getUserFn: at.getCorrectUser, wantTooManyRequests: true}, // too many requests
		{name: "too many requests", getUserFn: at.getCorrectUser, wantTooManyRequests: true}, // too many requests #2
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			req, cancel := newRequest(t, GET, url, nil)
			defer cancel()

			for _, c := range jar.Cookies(req.URL) {
				req.AddCookie(c)
			}

			if tt.getUserFn != nil {
				var (
					username = tt.getUserFn()
					pass     = ""
				)
				if tt.getPassFn != nil {
					pass = tt.getPassFn()
				}
				req.SetBasicAuth(username, pass)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(err)
			defer resp.Body.Close()

			var baseResp models.BaseResponse
			dec := json.NewDecoder(resp.Body)
			require.NoError(dec.Decode(&baseResp))
			require.False(dec.More())

			var (
				wantStatusCode         = http.StatusUnauthorized
				wantError              = "unauthorized"
				wantAuthenticateHeader = `Basic realm="Budget Manager"`
			)
			switch {
			case tt.wantAuthorized:
				wantStatusCode = http.StatusOK
				wantError = ""
				wantAuthenticateHeader = ""
			case tt.wantTooManyRequests:
				wantStatusCode = http.StatusTooManyRequests
				wantError = "too many requests"
				wantAuthenticateHeader = ""
			}

			require.Equal(wantStatusCode, resp.StatusCode)
			require.Equal(wantError, baseResp.Error)
			require.Equal(wantAuthenticateHeader, resp.Header.Get("WWW-Authenticate"))

			jar.SetCookies(resp.Request.URL, resp.Cookies())
		})
	}
}
