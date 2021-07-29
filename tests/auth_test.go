package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
)

func testAuth(t *testing.T, cfg app.Config) {
	url := fmt.Sprintf("http://localhost:%d/api/search/spends", cfg.Server.Port)

	tests := []struct {
		name               string
		username, password string
		wantStatusCode     int
	}{
		{
			name:           "no auth",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "wrong credentials",
			username:       "123",
			password:       "123",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "correct credentials",
			username:       "user",
			password:       "qwerty",
			wantStatusCode: http.StatusOK,
		},
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
			require.Equal(tt.wantStatusCode, resp.StatusCode)

			resp.Body.Close()
		})
	}
}
