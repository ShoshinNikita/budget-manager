package middlewares

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLimiter(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	const (
		firstIP  = "1"
		secondIP = "2"
	)
	forIPs := func(f func(ip string)) {
		f(firstIP)
		f(secondIP)
	}

	limiterCfg := rateLimiterConfig{
		interval:        100 * time.Millisecond,
		burst:           2,
		maxAge:          500 * time.Millisecond,
		cleanupInterval: 250 * time.Millisecond,
	}
	limiter := newRateLimiter(limiterCfg)

	// Check burst and reset
	forIPs(func(ip string) {
		for i := 0; i < 2; i++ {
			if i > 0 {
				limiter.reset(ip)
			}

			for j := 0; j < limiterCfg.burst; j++ {
				require.True(limiter.allow(ip))
			}
			require.False(limiter.allow(ip))
		}
	})

	// Add 1 token
	time.Sleep(limiterCfg.interval)
	forIPs(func(ip string) {
		require.True(limiter.allow(ip))
		require.False(limiter.allow(ip))
	})

	// Add 2 tokens
	time.Sleep(2 * limiterCfg.interval)
	forIPs(func(ip string) {
		require.True(limiter.allow(ip))
		require.True(limiter.allow(ip))
		require.False(limiter.allow(ip))
	})

	// Remove a limiter for 2nd IP by cleanup process
	time.Sleep(limiterCfg.maxAge / 2)
	require.True(limiter.allow(firstIP))
	time.Sleep(limiterCfg.maxAge / 2)
	require.True(limiter.allow(firstIP))
	time.Sleep(limiterCfg.maxAge / 2)

	require.Contains(limiter.limiters, firstIP)
	require.NotContains(limiter.limiters, secondIP)
}

func TestGetIP(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		remoteAddr string
		headers    map[string]string
		//
		want string
	}{
		{
			remoteAddr: "localhost",
			headers:    nil,
			want:       "localhost",
		},
		{
			remoteAddr: "127.0.0.1:1234",
			headers:    nil,
			want:       "127.0.0.1",
		},
		{
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "1.1.1.1"},
			want:       "1.1.1.1",
		},
		{
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "1.0.0.1, 192.168.0.10"},
			want:       "1.0.0.1",
		},
		{
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "8.8.8.8"},
			want:       "8.8.8.8",
		},
	} {
		tt := tt
		t.Run("", func(t *testing.T) {
			req := httptest.NewRequest("", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getIP(req)
			require.Equal(t, tt.want, got)
		})
	}
}
