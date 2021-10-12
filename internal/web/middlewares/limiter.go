package middlewares

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type rateLimiter struct {
	cfg      rateLimiterConfig
	mu       sync.Mutex
	limiters map[string]limiter
}

type rateLimiterConfig struct {
	interval        time.Duration
	burst           int
	maxAge          time.Duration
	cleanupInterval time.Duration
}

func newRateLimiter(cfg rateLimiterConfig) *rateLimiter {
	rl := &rateLimiter{
		cfg:      cfg,
		limiters: make(map[string]limiter),
	}
	rl.startCleanup()
	return rl
}

type limiter struct {
	limiter *rate.Limiter
	lastUse time.Time
}

func (rl *rateLimiter) Allow(r *http.Request) bool {
	return rl.allow(getIP(r))
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	lim, ok := rl.limiters[ip]
	if !ok {
		lim = limiter{
			limiter: rate.NewLimiter(rate.Every(rl.cfg.interval), rl.cfg.burst),
		}
	}
	lim.lastUse = time.Now()

	allow := lim.limiter.Allow()

	rl.limiters[ip] = lim

	return allow
}

func (rl *rateLimiter) Reset(r *http.Request) {
	rl.reset(getIP(r))
}

func (rl *rateLimiter) reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.limiters, ip)
}

func (rl *rateLimiter) startCleanup() {
	// TODO: add shutdown signal?
	go func() {
		for range time.Tick(rl.cfg.cleanupInterval) {
			rl.cleanup(time.Now())
		}
	}()
}

func (rl *rateLimiter) cleanup(now time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for k, v := range rl.limiters {
		if now.Sub(v.lastUse) >= rl.cfg.maxAge {
			delete(rl.limiters, k)
		}
	}
}

func getIP(r *http.Request) string {
	if h := r.Header.Get("X-Real-IP"); h != "" {
		return h
	}
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		if ips := strings.SplitN(h, ",", 2); len(ips) > 1 {
			return strings.TrimSpace(ips[0])
		}
		return h
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
