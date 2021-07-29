package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	cfg := app.Config{
		DBType:     "postgres",
		PostgresDB: pg.Config{Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
		Server: web.Config{
			UseEmbed: true,
			SkipAuth: false,
			Credentials: web.Credentials{
				"user": "$apr1$cpHMFyv.$BSB0aaF3bOrTC2f3V2VYG/", // user:qwerty
			},
			EnableProfiling: false,
		},
	}
	prepareApp(t, &cfg, StartPostgreSQL)

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
