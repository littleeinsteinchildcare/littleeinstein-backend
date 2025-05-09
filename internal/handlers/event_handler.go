package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/handlers"
	"net/http"
	"strings"
)

// EventService interface implemented in services package
type EventService interface {
	CreateEvent(user models.Event) error
	GetEventByID(id string) (models.Event, error)
	DeleteEventByID(id string) (bool, error)
	UpdateEvent(newData models.Event) (models.Event, error)
}

// EventHandler handles HTTP requests related to users
type EventHandler struct {
	eventService EventService
	userService  UserService
}

// NewEventHandler creates a new event handler
func NewEventHandler(s EventService, us UserService) *EventHandler {
	return &EventHandler{
		eventService: s,
		userService:  us,
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

	response := map[string]interface{}{
		"id":        id,
		"eventname": event.EventName,
		"date":      event.Date,
		"starttime": event.StartTime,
		"endtime":   event.EndTime,
		"creator":   event.Creator,
		"invitees":  event.Invitees,
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

	creator, err := h.userService.GetUserByID(eventData["creator"].(string))
	// invitee_ids := eventData["invitees"].(string)
	invitee_ids := strings.Split(eventData["invitees"].(string), ",")
	var invitees_list []models.User

	for _, id := range invitee_ids {
		user, err := h.userService.GetUserByID(id)
		Handle(err)
		invitees_list = append(invitees_list, user)
	}

	event := models.Event{
		ID:        eventData["id"].(string),
		EventName: eventData["eventname"].(string),
		Date:      eventData["date"].(string),
		StartTime: eventData["starttime"].(string),
		EndTime:   eventData["endtime"].(string),
		Creator:   creator,
		Invitees:  invitees_list,
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
		"invitees":  event.Invitees,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	success := true
	msg := "Event Update Successfully"
	eventData, err := DecodeEventRequest(r)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	newData, err := io.ReadAll(r.Body)
	if err != nil {
		handlers.WriteJSONError(w, http.StatusBadRequest, "UserHandler.UpdateUser: Failed to read request body", err)
	}

	defer r.Body.Close()

	var event models.Event
	if err := 

	creator, err := h.userService.GetUserByID(eventData["creator"].(string))
	// invitee_ids := eventData["invitees"].(string)
	invitee_ids := strings.Split(eventData["invitees"].(string), ",")
	var invitees_list []models.User

	for _, id := range invitee_ids {
		user, err := h.userService.GetUserByID(id)
		Handle(err)
		invitees_list = append(invitees_list, user)
	}

	event := models.Event{
		ID:        eventData["id"].(string),
		EventName: eventData["eventname"].(string),
		Date:      eventData["date"].(string),
		StartTime: eventData["starttime"].(string),
		EndTime:   eventData["endtime"].(string),
		Creator:   creator,
		Invitees:  invitees_list,
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
		"invitees":  event.Invitees,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	success, err := h.eventService.DeleteEventByID(id)

	if err != nil {
		fmt.Printf("Error deleting event: %v", err)
	}

	response := map[string]interface{}{
		"id":      id,
		"success": success,
	}

	w.Header().Set("Content-Type", "application/json")
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
