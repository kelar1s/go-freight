package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func New() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				status := strconv.Itoa(ww.Status())
				method := r.Method

				route := "unknown"
				if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
					if pattern := routeCtx.RoutePattern(); pattern != "" {
						route = pattern
					}
				}

				if ww.Status() == http.StatusNotFound {
					route = "404_not_found"
				}

				httpRequestsTotal.WithLabelValues(method, route, status).Inc()
				httpRequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
