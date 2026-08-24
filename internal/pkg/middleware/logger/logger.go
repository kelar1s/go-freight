package logger

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type loggerKey struct{}

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	log = log.With(slog.String("component", "middleware/logger"))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			spanContext := trace.SpanContextFromContext(ctx)

			traceID := ""
			if spanContext.HasTraceID() {
				traceID = spanContext.TraceID().String()
			}

			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("trace_id", traceID),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				args := []any{
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", time.Since(t1)),
				}

				switch status := ww.Status(); {
				case status >= 500:
					entry.ErrorContext(ctx, "request completed with server error", args...)
				case status >= 400:
					entry.WarnContext(ctx, "request completed with client error", args...)
				default:
					entry.InfoContext(ctx, "request completed", args...)
				}
			}()

			ctxWithLog := context.WithValue(ctx, loggerKey{}, entry)
			next.ServeHTTP(ww, r.WithContext(ctxWithLog))
		})
	}
}

func FromContext(ctx context.Context, defaultLogger *slog.Logger) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return defaultLogger
}
