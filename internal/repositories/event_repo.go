package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/handlers"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"os"
	"strings"

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

	fmt.Printf("INSIDE REPO CREATE EVENT\n")

	var invitee_ids []string
	for _, user := range event.Invitees {
		invitee_ids = append(invitee_ids, user.ID)
	}
	ids_string := strings.Join(invitee_ids, ",")

	//https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
	eventEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: "Events",
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

	fmt.Printf("AFTER CREATING ENTITY: Entity is %v\n", eventEntity)

	//https://pkg.go.dev/encoding/json
	serializedEntity, err := json.Marshal(eventEntity)
	handlers.Handle(err)

	_, err = repo.serviceClient.CreateTable(context.Background(), tableName, nil)

	tableClient := repo.serviceClient.NewClient(tableName)

	_, err = tableClient.AddEntity(context.Background(), serializedEntity, nil)
	if err != nil {
		fmt.Printf("ERR IS NOT NIL\n")
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

	userRepoCfg := NewUserRepoConfig(os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"), os.Getenv("AZURE_STORAGE_SERVICE_URL"))
	userRepo := NewUserRepo(*userRepoCfg)

	creator, err := userRepo.GetUser("UsersTable", myEntity.Properties["Creator"].(string))

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

func (repo *EventRepository) DeleteEvent(tableName string, id string) (bool, error) {
	ctx := context.Background()
	pKey := "Events"
	tableClient := repo.serviceClient.NewClient(tableName)

	_, err := tableClient.DeleteEntity(ctx, pKey, id, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}
