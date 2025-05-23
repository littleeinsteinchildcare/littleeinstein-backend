package services

import (
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
	"strings"
)

const EVENTSTABLE = "EventsTable"

// EventRepo interface methods implemented in repositories package
type EventRepo interface {
	CreateEvent(tableName string, event models.Event) error
	GetEvent(tableName string, id string) (models.Event, error)
	GetAllEvents(tableName string) ([]models.EventEntity, error)
	DeleteEvent(tableName string, id string) error
	DeleteEventByUserID(tableName string, userID string) error
	RemoveInvitee(tableName string, userID string) error
	UpdateEvent(tableNAme string, model models.Event) (models.Event, error)
}

// EventService contains and handles a specific EventRepository object
type EventService struct {
	repo        EventRepo
	userService UserService
}

// NewEventService constructs and returns a EventService object
func NewEventService(r EventRepo, us UserService) *EventService {
	return &EventService{repo: r, userService: us}
}

// GetEventByID handles calling the EventRepository GetEvent function and returns the result of a query by the EventRepository
func (s *EventService) GetEventByID(id string) (models.Event, error) {

	event, err := s.repo.GetEvent(EVENTSTABLE, id)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	eventRows, err := s.repo.GetAllEvents(EVENTSTABLE)
	if err != nil {
		return []models.Event{}, err
	}

	// Avoid re-querying if a user has already been found in a previous creator/invitee list query
	cachedUsers := make(map[string]models.User)
	getUser := func(id string) (models.User, error) {
		if u, ok := cachedUsers[id]; ok {
			return u, nil
		}
		u, err := s.userService.GetUserByID(id)
		if err != nil {
			return models.User{}, err
		}
		cachedUsers[id] = u
		return u, nil
	}

	events := make([]models.Event, 0, len(eventRows))

	// For each row returned by repo call, get full Event data (Creator as User and Invitees as User slice)
	for _, r := range eventRows {
		creator, err := getUser(r.CreatorID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get Creator %s: %w", r.CreatorID, err)
		}

		invitee_ids := strings.Split(r.InviteeIDs, ",")
		invitees := make([]models.User, 0, len(invitee_ids))
		for _, id := range invitee_ids {
			u, err := getUser(id)
			if err != nil {
				return nil, fmt.Errorf("Failed to get invitee %s: %w", id, err)
			}
			invitees = append(invitees, u)
		}

		// Build list of all events in table
		events = append(events, models.Event{
			ID:        r.RowKey,
			EventName: r.EventName,
			Date:      r.Date,
			StartTime: r.StartTime,
			EndTime:   r.EndTime,
			Creator:   creator,
			Invitees:  invitees,
		})
	}

	return events, nil
}

// GetEventsByUser returns events where the user is either creator or invitee
func (s *EventService) GetEventsByUser(userId string) ([]models.Event, error) {
	allEvents, err := s.GetAllEvents()
	if err != nil {
		return []models.Event{}, err
	}
	
	var userEvents []models.Event
	for _, event := range allEvents {
		// Check if user is creator
		if event.Creator.ID == userId {
			userEvents = append(userEvents, event)
			continue
		}
		
		// Check if user is in invitees
		for _, invitee := range event.Invitees {
			if invitee.ID == userId {
				userEvents = append(userEvents, event)
				break
			}
		}
	}
	
	return userEvents, nil
}

// CreateEvent returns an error on a failed EventRepo call
func (s *EventService) CreateEvent(event models.Event) error {
	err := s.repo.CreateEvent(EVENTSTABLE, event)
	if err != nil {
		return err
	}
	return nil
}

// Update Event and handle errors from Event Repo
func (s *EventService) UpdateEvent(event models.Event) (models.Event, error) {
	event, err := s.repo.UpdateEvent(EVENTSTABLE, event)
	if err != nil {
		return event, err
	}
	return event, nil
}

// Remove an Event and handle errors from Event Repo
func (s *EventService) DeleteEventByID(id string) error {
	err := s.repo.DeleteEvent(EVENTSTABLE, id)
	if err != nil {
		return err
	}
	return nil
}
