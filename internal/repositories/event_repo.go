package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

const PKey = "Events"

// EventRepo handles Database access
type EventRepository struct {
	serviceClient aztables.ServiceClient
}

// NewEventRepo creates and returns a new, unconnected EventRepo object
func NewEventRepo(cfg config.AzTableConfig) services.EventRepo {
	cred, err := aztables.NewSharedKeyCredential(cfg.AzureAccountName, cfg.AzureAccountKey)
	handlers.Handle(err)
	client, err := aztables.NewServiceClientWithSharedKey(cfg.AzureContainerName, cred, nil)
	handlers.Handle(err)
	return &EventRepository{serviceClient: *client}
}

type eventEntity struct {
	aztables.Entity
	EventName  string `aztable:"EventName"`
	Date       string `aztable:"Date"`
	StartTime  string `aztable:"StartTime"`
	EndTime    string `aztable:"EndTime"`
	CreatorID  string `aztable:"Creator"`
	InviteeIDs string `aztable:"Invitees"`
}

// GetEvent retrieves and stores entity data in a Repo object
func (repo *EventRepository) GetEvent(tableName string, id string) (models.Event, error) {

	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	resp, err := tableClient.GetEntity(ctx, PKey, id, nil)
	if err != nil {
		return models.Event{}, err
	}

	var myEntity aztables.EDMEntity
	err = json.Unmarshal(resp.Value, &myEntity)
	handlers.Handle(err)

	cfg, err := config.LoadAzTableConfig()
	handlers.Handle(err)
	userRepo, err := NewUserRepo(*cfg)

	creator, err := userRepo.GetUser("UsersTable", myEntity.Properties["Creator"].(string))
	handlers.Handle(err)

	invitee_ids := strings.Split(myEntity.Properties["Invitees"].(string), ",")

	for _, id := range invitee_ids {
		fmt.Printf("%v ", id)
	}

	var invitees_list []models.User

	for _, id := range invitee_ids {
		user, err := userRepo.GetUser("UsersTable", id)
		handlers.Handle(err)
		invitees_list = append(invitees_list, user)
	}

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

func (repo *EventRepository) GetAllEvents(tableName string) ([]eventEntity, error) {

	//TODO -- Continue working with DTO Approach

	tableClient := repo.serviceClient.NewClient(tableName)
	filter := "PartitionKey eq 'Events'"
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
	}
	var events []eventEntity

	pager := tableClient.NewListEntitiesPager(options)
	pageCount := 0
	for pager.More() {
		response, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("EventRepo.GetAllEvents: Failed to acquire next page: %w", err)
		}
		pageCount += 1

		for _, tableData := range response.Entities {
			var entityData eventEntity
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
	handlers.Handle(err)

	_, err = repo.serviceClient.CreateTable(context.Background(), tableName, nil)

	tableClient := repo.serviceClient.NewClient(tableName)

	_, err = tableClient.AddEntity(context.Background(), serializedEntity, nil)
	if err != nil {
		return err
	}
	return nil
}

func (repo *EventRepository) UpdateEvent(tableName string, newEventData models.Event) (models.Event, error) {

	tableClient := repo.serviceClient.NewClient(tableName)
	event, err := repo.GetEvent(tableName, newEventData.ID)
	if err != nil {
		return models.Event{}, fmt.Errorf("EVentRepo.UpdateEvent: Failed to retrieve event ID %s from %s: %w", newEventData.ID, tableName, err)
	}
	err = event.Update(newEventData)
	if err != nil {
		return models.Event{}, fmt.Errorf("EventRepo.UpdateEvent: Failed to updated event ID %s's fields: %w", event.ID, err)
	}
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

func (repo *EventRepository) DeleteEvent(tableName string, id string) error {
	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	_, err := tableClient.DeleteEntity(ctx, PKey, id, nil)
	if err != nil {
		return err
	}

	return nil
}
