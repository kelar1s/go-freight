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
	return slog.String("error", err.Error())
}
