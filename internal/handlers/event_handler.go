package handlers

import (
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
	"net/http"
)

// EventService interface implemented in services package
type EventService interface {
	CreateEvent(user models.Event) error
	GetEventByID(id string) (models.Event, error)
}

// EventHandler handles HTTP requests related to users
type EventHandler struct {
	eventService EventService
}

// NewEventHandler creates a new event handler
func NewEventHandler(s EventService) *EventHandler {
	return &EventHandler{
		eventService: s,
	}
}

// GetEvent handles GET requests for a specific event
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from request
	id := r.PathValue("id")

	event, err := h.eventService.GetEventByID(id)
	if err != nil {
		fmt.Printf("Error retrieving user: %v", err)
	}

	fmt.Print("IN GETEVENT\n")
	response := map[string]interface{}{
		"id":        id,
		"eventname": event.EventName,
		"date":      event.Date,
		"starttime": event.StartTime,
		"endtime":   event.EndTime,
		"creator":   event.Creator,
		"Invitees":  event.Invitees,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUser handles POST requests to create a new user
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	success := true
	msg := "Event Created Successfully"
	eventData, err := DecodeEventRequest(r)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	event := models.Event{
		ID:        eventData["id"].(string),
		EventName: eventData["eventname"].(string),
		Date:      eventData["date"].(string),
		StartTime: eventData["starttime"].(string),
		EndTime:   eventData["endtime"].(string),
		Creator:   eventData["creator"].(models.User),
		Invitees:  eventData["Invitees"].([]models.User),
	}

	err = h.eventService.CreateEvent(event)
	if err != nil {
		msg = fmt.Sprintf("Error Creating Event: %v\n", err)
		success = false
	}

	response := map[string]interface{}{
		"success":   success,
		"message":   msg,
		"id":        event.ID,
		"eventname": event.EventName,
		"date":      event.Date,
		"starttime": event.StartTime,
		"endtime":   event.EndTime,
		"creator":   event.Creator,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func DecodeEventRequest(r *http.Request) (map[string]interface{}, error) {
	var eventData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&eventData)
	if err != nil {
		fmt.Printf("Failed to decode json request")
		return eventData, err
	}
	return eventData, nil
}
