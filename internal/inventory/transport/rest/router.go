package rest

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"

	_ "github.com/kelar1s/go-freight/docs"
	mwLogger "github.com/kelar1s/go-freight/internal/pkg/middleware/logger"
	mwMetrics "github.com/kelar1s/go-freight/internal/pkg/middleware/metrics"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	productHandler *ProductHandler,
	warehouseHandler *WarehouseHandler,
	log *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Get("/metrics", promhttp.Handler().ServeHTTP)
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(mwMetrics.New())

		r.Use(otelchi.Middleware("inventory-service", otelchi.WithChiRoutes(r)))

		r.Use(mwLogger.New(log))
		r.Use(middleware.Recoverer)

		r.Use(middleware.RealIP)
		r.Use(middleware.URLFormat)

		r.Route("/api/v1", func(r chi.Router) {
			r.Route("/warehouses", func(r chi.Router) {
				r.Post("/", warehouseHandler.Create)
				r.Get("/", warehouseHandler.List)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", warehouseHandler.Get)
					r.Put("/", warehouseHandler.Update)
					r.Delete("/", warehouseHandler.Delete)

					r.Get("/products", productHandler.ListByWarehouse)
				})
			})

			r.Route("/products", func(r chi.Router) {
				r.Post("/", productHandler.Create)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", productHandler.Get)
					r.Delete("/", productHandler.Delete)

					r.Patch("/adjust", productHandler.AdjustQuantity)
					r.Patch("/reserve", productHandler.Reserve)
					r.Patch("/release", productHandler.Release)
					r.Patch("/cancel-reservation", productHandler.CancelReservation)
				})
			})
		})
	})

	return r
}
