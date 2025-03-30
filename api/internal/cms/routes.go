package cms

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Route("/cms", func(r chi.Router) {
		r.Get("/health", handler.HealthCheckHandler)
	})
}
