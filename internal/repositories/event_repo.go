package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

// Current PartitionKey for events table
// ? Change this to hold the creatorID per event instead?
const PKey = "Events"

// EventRepo handles Database access
type EventRepository struct {
	serviceClient aztables.ServiceClient
}

// NewEventRepo creates and returns a new, unconnected EventRepo object
func NewEventRepo(cfg config.AzTableConfig) (services.EventRepo, error) {
	cred, err := aztables.NewSharedKeyCredential(cfg.AzureAccountName, cfg.AzureAccountKey)
	if err != nil {
		return nil, fmt.Errorf("EventRepo.NewEventRepo: Failed to create credentials: %w", err)
	}
	client, err := aztables.NewServiceClientWithSharedKey(cfg.AzureContainerName, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("EventRepo.NewEventRepo: Failed to initialize service client: %w", err)
	}
	return &EventRepository{serviceClient: *client}, nil
}

// GetEvent retrieves and stores entity data in a Repo object
func (repo *EventRepository) GetEvent(tableName string, id string) (models.Event, error) {

	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	// Get Event in EDMEntity form
	resp, err := tableClient.GetEntity(ctx, PKey, id, nil)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to retrieve entity from %s: %w", tableName, err)
	}

	// Deserialize data retrieved from table
	var myEntity aztables.EDMEntity
	err = json.Unmarshal(resp.Value, &myEntity)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to deserialize entity: %w", err)
	}

	//! START Hacky { - Create a new user repository to access user data
	//! If time allows - big point to refactor here using the DTO approach
	cfg, err := config.LoadAzTableConfig()
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to Load Aztable config %w", err)
	}
	userRepo, err := NewUserRepo(*cfg)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to initialize User Repo %w", err)
	}

	// Grab creator from User Repo to fill in the Event struct
	creator, err := userRepo.GetUser("UsersTable", myEntity.Properties["Creator"].(string))
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to get Creator %w", err)
	}

	// Process Invitee IDs to store list of Users in Event struct
	invitee_ids := strings.Split(myEntity.Properties["Invitees"].(string), ",")

	var invitees_list []models.User

	for _, id := range invitee_ids {
		user, err := userRepo.GetUser("UsersTable", id)
		if err != nil {
			return models.Event{}, fmt.Errorf("EventRepo.GetEvent: Failed to get invitee %w", err)
		}
		invitees_list = append(invitees_list, user)
	}

	//! END Hacky }

	event := models.Event{
		ID:        myEntity.RowKey,
		EventName: myEntity.Properties["EventName"].(string),
		Date:      myEntity.Properties["Date"].(string),
		StartTime: myEntity.Properties["StartTime"].(string),
		EndTime:   myEntity.Properties["EndTime"].(string),
		Creator:   creator,
		Invitees:  invitees_list,
	}

	return event, nil
}

func (repo *EventRepository) GetAllEvents(tableName string) ([]models.EventEntity, error) {

	// Establish new client and set filter
	tableClient := repo.serviceClient.NewClient(tableName)
	filter := "PartitionKey eq 'Events'"
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
	}
	var events []models.EventEntity

	// Iterate through all pages in table
	pager := tableClient.NewListEntitiesPager(options)
	pageCount := 0
	for pager.More() {
		response, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("EventRepo.GetAllEvents: Failed to acquire next page: %w", err)
		}
		pageCount += 1

		// After getting all the responses, deserialize and create list of EventEntities to be processed by service layer
		for _, tableData := range response.Entities {
			var entityData models.EventEntity
			err = json.Unmarshal(tableData, &entityData)
			if err != nil {
				return nil, fmt.Errorf("EventRepo.GetAllEvents: Failed to unmarshal entity: %w", err)
			}
			events = append(events, entityData)
		}
	}
	return events, nil
}

// CreateEvent creates an aztable entity in the specified table name, creating the table if it doesn't exist
func (repo *EventRepository) CreateEvent(tableName string, event models.Event) error {

	// Create CSV of Invitee IDs
	var invitee_ids []string
	for _, user := range event.Invitees {
		invitee_ids = append(invitee_ids, user.ID)
	}
	ids_string := strings.Join(invitee_ids, ",")

	//https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
	eventEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: PKey,
			RowKey:       event.ID,
		},
		Properties: map[string]any{
			"EventName": event.EventName,
			"Date":      event.Date,
			"StartTime": event.StartTime,
			"EndTime":   event.EndTime,
			"Creator":   event.Creator.ID,
			"Invitees":  ids_string,
		},
	}

	//https://pkg.go.dev/encoding/json
	serializedEntity, err := json.Marshal(eventEntity)
	if err != nil {
		return fmt.Errorf("EventRepo.CreateEvent: Failed to serialize entity %w", err)
	}

	// Create Table if for some reason it doesn't exist, no need to check error since we expect an error to get generated
	// TODO: Handle the case where the table doesn't exist? May need to explicity set row types again
	_, err = repo.serviceClient.CreateTable(context.Background(), tableName, nil)

	tableClient := repo.serviceClient.NewClient(tableName)

	_, err = tableClient.AddEntity(context.Background(), serializedEntity, nil)
	if err != nil {
		return fmt.Errorf("EventRepo.CreateEvent: Failed to add event entity %w", err)
	}

	return nil
}

// Update Event with partial or full updates (excluding Creator ID)
func (repo *EventRepository) UpdateEvent(tableName string, newEventData models.Event) (models.Event, error) {

	tableClient := repo.serviceClient.NewClient(tableName)
	event, err := repo.GetEvent(tableName, newEventData.ID)
	if err != nil {
		return models.Event{}, fmt.Errorf("EVentRepo.UpdateEvent: Failed to retrieve event ID %s from %s: %w", newEventData.ID, tableName, err)
	}
	// Update event fields inside Event object
	err = event.Update(newEventData)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.UpdateEvent: Failed to updated event ID %s's fields: %w", event.ID, err)
	}

	// Process Invitee IDs into CSV
	var invitee_ids []string
	for _, user := range event.Invitees {
		invitee_ids = append(invitee_ids, user.ID)
	}
	ids_string := strings.Join(invitee_ids, ",")

	eventEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: PKey,
			RowKey:       event.ID,
		},
		Properties: map[string]any{
			"EventName": event.EventName,
			"Date":      event.Date,
			"StartTime": event.StartTime,
			"EndTime":   event.EndTime,
			"Creator":   event.Creator.ID,
			"Invitees":  ids_string,
		},
	}

	// Serialize and Update
	serEntity, err := json.Marshal(eventEntity)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.UpdateUser: Failed to serialize event data: %w", err)
	}
	_, err = tableClient.UpdateEntity(context.Background(), serEntity, nil)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.UpdateEntity: Failed to update entity in %s: %w", tableName, err)
	}

	return event, nil
}

// Remove an Event from the table based on Event ID
func (repo *EventRepository) DeleteEvent(tableName string, id string) error {
	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	_, err := tableClient.DeleteEntity(ctx, PKey, id, nil)
	if err != nil {
		return fmt.Errorf("EventRepo.DeleteEvent: Failed to delete entity in %s: %w", tableName, err)
	}

	return nil
}

// Remove all Events from Events Table based on Creator ID
func (repo *EventRepository) DeleteEventByUserID(tableName string, userID string) error {

	tableClient := repo.serviceClient.NewClient(tableName)
	filter := fmt.Sprintf("Creator eq '%s'", userID)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
	}

	// Iterate through rows, removing all events whose Creator ID matches the UserID argument
	pager := tableClient.NewListEntitiesPager(options)
	pageCount := 0
	for pager.More() {
		response, err := pager.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("EventRepo.DeleteEventByUserID: Failed to acquire next page: %w", err)
		}
		pageCount += 1

		for _, tableData := range response.Entities {
			var entityData models.EventEntity
			err = json.Unmarshal(tableData, &entityData)
			repo.DeleteEvent(services.EVENTSTABLE, entityData.CreatorID)
			if err != nil {
				return fmt.Errorf("EventRepo.DeleteEventByUserID: Failed to unmarshal entity: %w", err)
			}
		}
	}
	return nil
}

// Remove Invitee from all Events in Events Table based on Invitee ID
func (repo *EventRepository) RemoveInvitee(tableName string, userID string) error {

	tableClient := repo.serviceClient.NewClient(tableName)

	// Iterate through rows, updating the Invitees list by removing the specified UserID
	pager := tableClient.NewListEntitiesPager(nil)
	pageCount := 0
	for pager.More() {
		response, err := pager.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("EventRepo.DeleteEventByUserID: Failed to acquire next page: %w", err)
		}
		pageCount += 1

		for _, tableData := range response.Entities {
			var entityData models.EventEntity
			err = json.Unmarshal(tableData, &entityData)
			// Strip UserID, update Event Entity
			if strings.Contains(entityData.InviteeIDs, userID) {
				updatedInvitees := repo.stripInvitee(entityData.InviteeIDs, userID)
				entityData.InviteeIDs = updatedInvitees

				serEntity, err := json.Marshal(entityData)
				if err != nil {
					return fmt.Errorf("EventRepo.RemoveInvitee: Failed to serialize event data: %w", err)
				}
				_, err = tableClient.UpdateEntity(context.Background(), serEntity, nil)
				if err != nil {
					return fmt.Errorf("EventRepo.RemoveInvitee: Failed to update entity in %s: %w", tableName, err)
				}
			}
		}
	}
	return nil
}

// Helper function to process CSV by removing specified token
func (repo *EventRepository) stripInvitee(csv, toRemove string) string {
	tokens := strings.Split(csv, ",")
	filtered := make([]string, 0, len(tokens))
	for _, p := range tokens {
		if p != toRemove {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, ",")
}