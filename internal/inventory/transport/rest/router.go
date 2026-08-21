package rest

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/kelar1s/go-freight/docs"
	mwLogger "github.com/kelar1s/go-freight/internal/pkg/middleware/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	productHandler *ProductHandler,
	warehouseHandler *WarehouseHandler,
	log *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.RedirectSlashes)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/warehouses", func(r chi.Router) {
			r.Post("", warehouseHandler.Create)
			r.Get("", warehouseHandler.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("", warehouseHandler.Get)
				r.Put("", warehouseHandler.Update)
				r.Delete("", warehouseHandler.Delete)

				r.Get("/products", productHandler.ListByWarehouse)
			})
		})

		r.Route("/products", func(r chi.Router) {
			r.Post("", productHandler.Create)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("", productHandler.Get)
				r.Delete("", productHandler.Delete)

				r.Patch("/adjust", productHandler.AdjustQuantity)
				r.Patch("/reserve", productHandler.Reserve)
				r.Patch("/release", productHandler.Release)
				r.Patch("/cancel-reservation", productHandler.CancelReservation)
			})
		})
	})

	return r
}
