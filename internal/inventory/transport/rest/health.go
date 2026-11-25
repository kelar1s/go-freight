package rest

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type ReadinessChecker interface {
	CheckReady(ctx context.Context) error
}

type HealthHandler struct {
	checker ReadinessChecker
	log     *slog.Logger
}

func NewHealthHandler(checker ReadinessChecker, log *slog.Logger) *HealthHandler {
	return &HealthHandler{
		checker: checker,
		log:     log,
	}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.checker == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.checker.CheckReady(ctx); err != nil {
		h.log.WarnContext(r.Context(), "service is not ready", slog.String("error", err.Error()))
		WriteError(w, http.StatusServiceUnavailable, "service not ready")
		return
	}

	w.WriteHeader(http.StatusOK)
}
