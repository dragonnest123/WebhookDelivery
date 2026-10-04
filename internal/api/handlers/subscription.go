package handlers

import (
	"encoding/json"
	"net/http"
	jsonHelper "webhookdelivery/internal/api/helpers"
	"webhookdelivery/internal/db/repositories"
	sqlc "webhookdelivery/internal/db/sqlc"
)

type SubscriptionHandler struct {
	repository repositories.SubscriptionRepository
}

func NewSubscriptionHandler(repository repositories.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{
		repository: repository,
	}
}

func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var params sqlc.CreateSubscriptionParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subscription, err := h.repository.CreateSubscription(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJSON(w, http.StatusCreated, subscription)
}
