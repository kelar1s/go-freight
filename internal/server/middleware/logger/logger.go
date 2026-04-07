package logger

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type loggerKey struct{}

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	log = log.With(slog.String("component", "middleware/logger"))
	log.Info("logger middleware enabled")
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				args := []any{
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", time.Since(t1)),
				}
				status := ww.Status()
				switch {
				case status >= 500:
					entry.Error("request completed with error: ", args...)
				case status >= 400:
					entry.Warn("request completed with client error", args...)
				default:
					entry.Info("request completed", args...)
				}
			}()

			ctxWithLog := context.WithValue(r.Context(), loggerKey{}, entry)
			r = r.WithContext(ctxWithLog)

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func FromContext(ctx context.Context, defaultLogger *slog.Logger) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return defaultLogger
}
