package handlers

import (
	"encoding/json"
	"net/http"
	jsonHelper "webhookdelivery/internal/api/helpers"
	sqlc "webhookdelivery/internal/db/sqlc"
	"webhookdelivery/internal/services"
)

type EventHandler struct {
	eventService services.EventService
}

func NewEventHandler(eventService services.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) SendEvent(w http.ResponseWriter, r *http.Request) {
	var params sqlc.CreateEventParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.CreateEvent(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJSON(w, http.StatusCreated, event)
}
