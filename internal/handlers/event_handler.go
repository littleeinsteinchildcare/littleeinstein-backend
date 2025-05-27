package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/utils"
	"net/http"
	"strings"
	"littleeinsteinchildcare/backend/internal/common"
)

// EventService interface implemented in services package
type EventService interface {
	CreateEvent(user models.Event) error
	GetEventByID(id string) (models.Event, error)
	GetAllEvents() ([]models.Event, error)
	GetEventsByUser(userId string) ([]models.Event, error)
	DeleteEventByID(id string) error
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
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("EventHandler.GetEvent: Failed to find User with ID %s", id), err)
	}

	response := buildEventResponse(event)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Return all events with full User data for Creator and Invitees
func (h *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventService.GetAllEvents()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, fmt.Sprintf("EventHandler.GetAllEvents: Failed to retrieve list of events"), err)
		return
	}
	var responses []map[string]interface{}

	for _, event := range events {
		resp := buildEventResponse(event)
		responses = append(responses, resp)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)

}

// Return events where user is creator or invitee
func (h *EventHandler) GetEventsByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	
	events, err := h.eventService.GetEventsByUser(userId)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("EventHandler.GetEventsByUser: Failed to retrieve events for user %s", userId), err)
		return
	}
	
	var responses []map[string]interface{}
	for _, event := range events {
		resp := buildEventResponse(event)
		log.Printf("DEBUG: Event in GetEventsByUser: ID=%s, Name=%s, Location=%s, Description=%s, Color=%s", 
			event.ID, event.EventName, event.Location, event.Description, event.Color)
		responses = append(responses, resp)
	}
	
	log.Printf("DEBUG: GetEventsByUser returning %d events to frontend", len(responses))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// CreateEvent handles POST requests to create a new user
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	eventData, err := utils.DecodeJSONRequest(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "CreateEvent.DecodeRequest: Failed to decode JSON request", err)
		return
	}

	creatorID, err := utils.GetUserIDFromAuth(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, fmt.Sprintf("EventHandler.CreateEvent: Failed to get user ID from auth: %v", err), err)
		return
	}
	log.Printf("DEBUG: Got creator ID from auth: %s", creatorID)
	
	// creator, err := h.userService.GetUserByID(eventData["creator"].(string))
	log.Printf("DEBUG: About to call userService.GetUserByID with ID: '%s'", creatorID)
	creator, err := h.userService.GetUserByID(creatorID)
	if err != nil {
		log.Printf("DEBUG: userService.GetUserByID failed with error: %v", err)
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("EventHandler.CreateEvent: Failed to find Creator with ID %s:", creatorID), err)
		return
	}
	log.Printf("DEBUG: Successfully retrieved creator: %+v", creator)
	inviteesStr := eventData["invitees"].(string)
	log.Printf("DEBUG: Processing invitees string: '%s'", inviteesStr)
	invitee_ids := strings.Split(inviteesStr, ",")
	log.Printf("DEBUG: Split invitee IDs: %+v", invitee_ids)
	var invitees_list []models.User

	for _, id := range invitee_ids {
		id = strings.TrimSpace(id)
		log.Printf("DEBUG: Processing invitee ID: '%s'", id)
		if id == "" {
			log.Printf("DEBUG: Skipping empty invitee ID")
			continue
		}
		user, err := h.userService.GetUserByID(id)
		if err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("EventHandler.CreateEvent: Failed to find User with ID %s", id), err)
			return
		}
		invitees_list = append(invitees_list, user)
	}

	// Extract additional fields with safe type assertions
	location := ""
	if loc, ok := eventData["location"].(string); ok {
		location = loc
	}
	
	description := ""
	if desc, ok := eventData["description"].(string); ok {
		description = desc
	}
	
	color := "#4CAF50" // Default green color
	if col, ok := eventData["color"].(string); ok && col != "" {
		color = col
	}

	event := models.Event{
		ID:          eventData["id"].(string),
		EventName:   eventData["eventname"].(string),
		Date:        eventData["date"].(string),
		StartTime:   eventData["starttime"].(string),
		EndTime:     eventData["endtime"].(string),
		Location:    location,
		Description: description,
		Color:       color,
		Creator:     creator,
		Invitees:    invitees_list,
	}
	
	log.Printf("DEBUG: Created event object: ID=%s, Name=%s, Location=%s, Description=%s, Color=%s", 
		event.ID, event.EventName, event.Location, event.Description, event.Color)

	err = h.eventService.CreateEvent(event)
	if err != nil {
		utils.WriteJSONError(w, http.StatusConflict, "EventHandler.CreateEvent: Failed to create Event", err)
		return
	}

	response := buildEventResponse(event)
	log.Printf("DEBUG: Event response being sent to frontend: %+v", response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Update Event for PUT requests, using partial matches to replace specified fields
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {

	pathID := r.PathValue("id")
	eventData, err := utils.DecodeJSONRequest(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "EventHandler.UpdateEvent: Failed to Decode JSON", nil)
		return
	}
	if _, ok := eventData["id"]; !ok {
		utils.WriteJSONError(w, http.StatusBadRequest, "EventHandler.UpdateEvent: Missing required field: id", nil)
		return
	}

	if eventData["id"].(string) != pathID {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("EventHandler.UpdateEvent: Path ID %s does not match JSON ID %s", pathID, eventData["id"]), errors.New("ID mismatch"))
		return
	}

	event, err := h.BuildPartialEvent(w, eventData)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("EventHandler.UpdateEvent: Failed to build partial event"), err)
		return
	}

	updatedEvent, err := h.eventService.UpdateEvent(event)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, "EventHandler.UpdateEvent: Event does not exist", err)
		return
	}

	response := buildEventResponse(updatedEvent)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Delete an Event by Event ID
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.eventService.DeleteEventByID(id)

	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("Error deleting event %s", id), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to create Event to pass to service layer
func (h *EventHandler) BuildPartialEvent(w http.ResponseWriter, eventData map[string]any) (models.Event, error) {

	event := models.Event{ID: eventData["id"].(string)} // Above handling assures this will always exist in the eventData

	if v, ok := eventData["eventname"].(string); ok {
		event.EventName = v
	}
	if v, ok := eventData["date"].(string); ok {
		event.Date = v
	}
	if v, ok := eventData["starttime"].(string); ok {
		event.StartTime = v
	}
	if v, ok := eventData["endtime"].(string); ok {
		event.EndTime = v
	}
	if v, ok := eventData["location"].(string); ok {
		event.Location = v
	}
	if v, ok := eventData["description"].(string); ok {
		event.Description = v
	}
	if v, ok := eventData["color"].(string); ok {
		event.Color = v
	}

	//! Questionable - Given time for a refactor, this could be cleaner with a better overall structure
	// Grab IDs from Event Data and populate Event object with relevant User objects
	var creator models.User
	var err error
	if eventData["creator"] != nil {
		creator, err = h.userService.GetUserByID(eventData["creator"].(string))
		if err != nil {
			return event, err
		}
		event.Creator = creator
	}

	var invitees_list []models.User
	if eventData["invitees"] != nil {
		invitee_ids := strings.Split(eventData["invitees"].(string), ",")

		for _, id := range invitee_ids {
			id = strings.TrimSpace(id)
			user, err := h.userService.GetUserByID(id)
			if err != nil {
				return event, err
			}
			invitees_list = append(invitees_list, user)
		}
		event.Invitees = invitees_list
	}
	return event, nil
}

// Helper functiont to package JSON response
func buildEventResponse(event models.Event) map[string]interface{} {
	response := map[string]interface{}{
		"id":          event.ID,
		"eventname":   event.EventName,
		"date":        event.Date,
		"starttime":   event.StartTime,
		"endtime":     event.EndTime,
		"location":    event.Location,
		"description": event.Description,
		"color":       event.Color,
		"creator":     event.Creator,
		"invitees":    event.Invitees,
	}
	return response
}

func (h *EventHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(common.ContextUID).(string)
	if !ok {
		http.Error(w, "UID missing in context", http.StatusInternalServerError)
		return
	}

	email, _ := r.Context().Value(common.ContextEmail).(string)

	response := map[string]interface{}{
		"status": "authenticated",
		"uid":    uid,
		"email":  email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}