package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type contextKey string

const (
	ContextKeyLogger    contextKey = "logger"
	ContextKeyRequestID contextKey = "request_id"
)

func WithLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger := logger.With(
				// zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			now := time.Now()

			defer func() {
				logger.Info("handled",
					zap.Int("status_code", ww.Status()),
					zap.Duration("duration", time.Since(now)),
				)
			}()

			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyLogger, logger)
			next.ServeHTTP(ww, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
