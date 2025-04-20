package repositories

import (
	"context"
	"encoding/json"
	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

// EventRepoConfig stores connection information to be passed in to the EventRepo constructor
type EventRepoConfig struct {
	accountName        string
	accountKey         string
	serviceEndpointURL string
}

// NewEventRepoConfig constructs a new EventEventConfig object and returns it
func NewEventRepoConfig(name string, key string, url string) *EventRepoConfig {
	return &EventRepoConfig{
		accountName:        name,
		accountKey:         key,
		serviceEndpointURL: url,
	}
}

// EventRepo handles Database access
type EventRepository struct {
	serviceClient aztables.ServiceClient
}

// NewEventRepo creates and returns a new, unconnected EventRepo object
func NewEventRepo(cfg EventRepoConfig) services.EventRepo {
	cred, err := aztables.NewSharedKeyCredential(cfg.accountName, cfg.accountKey)
	handlers.Handle(err)
	client, err := aztables.NewServiceClientWithSharedKey(cfg.serviceEndpointURL, cred, nil)
	handlers.Handle(err)
	return &EventRepository{serviceClient: *client}
}

// CreateEvent creates an aztable entity in the specified table name, creating the table if it doesn't exist
func (repo *EventRepository) CreateEvent(tableName string, event models.Event) error {

	//https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
	eventEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: "Events",
			RowKey:       event.ID,
		},
		Properties: map[string]any{
			"Eventname": event.EventName,
			"Date":      event.Date,
			"StartTime": event.StartTime,
			"EndTime":   event.EndTime,
			"Creator":   event.Creator,
			"Invitees":  event.Invitees,
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

// GetEvent retrieves and stores entity data in a Repo object
func (repo *EventRepository) GetEvent(tableName string, id string) (models.Event, error) {

	ctx := context.Background()
	pKey := "Events"
	tableClient := repo.serviceClient.NewClient(tableName)

	resp, err := tableClient.GetEntity(ctx, pKey, id, nil)
	if err != nil {
		return models.Event{}, err
	}

	var myEntity aztables.EDMEntity
	err = json.Unmarshal(resp.Value, &myEntity)
	handlers.Handle(err)

	event := models.Event{
		ID:        myEntity.RowKey,
		EventName: myEntity.Properties["Eventname"].(string),
		Date:      myEntity.Properties["Date"].(string),
		StartTime: myEntity.Properties["StartTime"].(string),
		EndTime:   myEntity.Properties["EndTime"].(string),
		Creator:   myEntity.Properties["Creator"].(models.User),
		Invitees:  myEntity.Properties["Invitees"].([]models.User),
	}

	return event, nil
}
