package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Mars-60/project4/backend/configs"
	"github.com/Mars-60/project4/backend/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewRouter(app *App) *chi.Mux {

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(securityHeaders)
	router.Use(defaultRateLimiter().Middleware)
	router.Use(requestLogger)
	router.Use(cors)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(configs.App.Server.WriteTimeout))

	RegisterRoutes(router, app)

	return router
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]rateWindow
}

type rateWindow struct {
	count   int
	resetAt time.Time
}

func defaultRateLimiter() *rateLimiter {
	return &rateLimiter{limit: 120, window: time.Minute, clients: make(map[string]rateWindow)}
}

func (l *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(r.RemoteAddr, time.Now()) {
			WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) allow(remoteAddr string, now time.Time) bool {
	host := remoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx > -1 {
		host = remoteAddr[:idx]
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.clients[host]
	if window.resetAt.IsZero() || now.After(window.resetAt) {
		l.clients[host] = rateWindow{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if window.count >= l.limit {
		return false
	}
	window.count++
	l.clients[host] = window
	return true
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			logger.Log.Info(
				"http request",
				zap.String("request_id", middleware.GetReqID(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", time.Since(startedAt)),
				zap.String("remote_addr", r.RemoteAddr),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

func cors(next http.Handler) http.Handler {
	allowedOrigins := toSet(configs.App.Server.CORS.AllowedOrigins)
	allowedMethods := strings.Join(configs.App.Server.CORS.AllowedMethods, ", ")
	allowedHeaders := strings.Join(configs.App.Server.CORS.AllowedHeaders, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (allowedOrigins["*"] || allowedOrigins[origin]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}

	return set
}
