package api

import (
	"webhookdelivery/internal/api/handlers"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	*chi.Mux
}

func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

func (r *Router) RegisterHealthHandler(h *handlers.HealthHandler) {
	r.Get("/health", h.CheckHealth)
}

func (r *Router) RegisterSubscriptionHandler(h *handlers.SubscriptionHandler) {
	r.Post("/subscriptions", h.CreateSubscription)
}

func (r *Router) RegisterEventHandler(h *handlers.EventHandler) {
	r.Post("/events", h.SendEvent)
}
