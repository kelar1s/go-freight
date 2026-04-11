package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kelar1s/go-freight/internal/inventory/handler"
	mwLogger "github.com/kelar1s/go-freight/internal/server/middleware/logger"
)

func NewRouter(h *handler.ProductHandler, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/warehouses", func(r chi.Router) {
			r.Post("/", h.CreateWarehouse)
			r.Get("/", h.ListWarehouses)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetWarehouse)
				r.Put("/", h.UpdateWarehouse)
				r.Delete("/", h.DeleteWarehouse)
				r.Get("/products", h.ListProductsByWarehouse)
			})
		})

		r.Route("/products", func(r chi.Router) {
			r.Post("/", h.CreateProduct)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetProduct)
				r.Delete("/", h.DeleteProduct)
				r.Patch("/add", h.AddProductQuantity)
				r.Patch("/set", h.SetProductQuantity)
			})
		})
	})

	return r
}
