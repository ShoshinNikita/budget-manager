package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/web"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	RunTest(t, TestFn(testAuth), func(env *TestEnv) {
		env.Cfg.Server.Auth.Disable = false
		env.Cfg.Server.Auth.BasicAuthCreds = web.Credentials{
			"user": "$2y$05$wK5Ad.qdY.ZLPsfEv3rc/.uO.8SkbD6r2ptiuZefMUOX0wgGK/1rC", // user:qwerty
		}
	})
}

func testAuth(t *testing.T, host string) {
	url := fmt.Sprintf("http://%s/api/search/spends", host)

	const (
		User  = "user"
		Pass  = "qwerty"
		Wrong = "123"
	)

	tests := []struct {
		name               string
		username, password string
		//
		wantAuthorized bool
	}{
		{name: "no auth"},
		{name: "wrong username and password", username: Wrong, password: Wrong},
		{name: "wrong username", username: Wrong, password: Pass},
		{name: "wrong password", username: User, password: Wrong},
		{name: "correct credentials", username: User, password: Pass, wantAuthorized: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			req, cancel := newRequest(t, GET, url, nil)
			defer cancel()

			if tt.username != "" {
				req.SetBasicAuth(tt.username, tt.password)
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
			if tt.wantAuthorized {
				wantStatusCode = http.StatusOK
				wantError = ""
				wantAuthenticateHeader = ""
			}

			require.Equal(wantStatusCode, resp.StatusCode)
			require.Equal(wantError, baseResp.Error)
			require.Equal(wantAuthenticateHeader, resp.Header.Get("WWW-Authenticate"))
		})
	}
}
