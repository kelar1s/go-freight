package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

type otelHandler struct {
	slog.Handler
}

func (h otelHandler) Handle(ctx context.Context, r slog.Record) error {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		r.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return otelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h otelHandler) WithGroup(name string) slog.Handler {
	return otelHandler{Handler: h.Handler.WithGroup(name)}
}

func Setup(env string) *slog.Logger {
	var baseHandler slog.Handler

	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	switch env {
	case envLocal:
		baseHandler = slog.NewTextHandler(os.Stdout, opts)
	case envDev:
		baseHandler = slog.NewJSONHandler(os.Stdout, opts)
	case envProd:
		opts.Level = slog.LevelInfo
		baseHandler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		opts.Level = slog.LevelInfo
		baseHandler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(otelHandler{Handler: baseHandler})
}

func Err(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.String("error", err.Error())
}
