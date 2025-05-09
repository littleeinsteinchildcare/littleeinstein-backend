package services

import (
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
)

const EVENTSTABLE = "EventsTable"

// EventRepo interface methods implemented in repositories package
type EventRepo interface {
	CreateEvent(tableName string, event models.Event) error
	GetEvent(tableName string, id string) (models.Event, error)
	DeleteEvent(tableName string, id string) (bool, error)
	UpdateEvent(tableNAme string, model models.Event) (models.Event, error)
}

// EventService contains and handles a specific EventRepository object
type EventService struct {
	repo EventRepo
}

// NewEventService constructs and returns a EventService object
func NewEventService(r EventRepo) *EventService {
	return &EventService{repo: r}
}

// GetEventByID handles calling the EventRepository GetEvent function and returns the result of a query by the EventRepository
func (s *EventService) GetEventByID(id string) (models.Event, error) {

	event, err := s.repo.GetEvent(EVENTSTABLE, id)
	if err != nil {
		return models.Event{}, err
	}

	fmt.Printf("Event: %v", event)
	return event, nil
}

// CreateEvent returns an error on a failed EventRepo call
func (s *EventService) CreateEvent(event models.Event) error {
	err := s.repo.CreateEvent(EVENTSTABLE, event)
	if err != nil {
		return err
	}
	return nil
}

func (s *EventService) DeleteEventByID(id string) (bool, error) {
	success, err := s.repo.DeleteEvent(EVENTSTABLE, id)
	if err != nil {
		return success, err
	}
	return success, nil
}
