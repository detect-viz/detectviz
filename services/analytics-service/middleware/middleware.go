package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Chain(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) == 0 {
		return h
	}

	wrapped := h
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}

	return wrapped
}

func Logger(logger *log.Logger) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Printf("Started %s %s", r.Method, r.URL.Path)
			next(w, r)
			logger.Printf("Completed in %v", time.Since(start))
		}
	}
}

func Timeout(timeout time.Duration) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan bool)

			go func() {
				next(w, r)
				done <- true
			}()

			select {
			case <-ctx.Done():
				w.WriteHeader(http.StatusGatewayTimeout)
				return
			case <-done:
				return
			}
		}
	}
}

func RateLimit(rps int) Middleware {
	limiter := rate.NewLimiter(rate.Limit(rps), rps)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}
