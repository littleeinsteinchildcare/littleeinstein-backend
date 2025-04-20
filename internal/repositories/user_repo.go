package repositories

import (
	"context"
	"encoding/json"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

// UserRepoConfig stores connection information to be passed in to the UserRepo constructor
type UserRepoConfig struct {
	accountName        string
	accountKey         string
	serviceEndpointURL string
}

// NewUserRepoConfig constructs a new UserRepoConfig object and returns it
func NewUserRepoConfig(name string, key string, url string) *UserRepoConfig {
	return &UserRepoConfig{
		accountName:        name,
		accountKey:         key,
		serviceEndpointURL: url,
	}
}

// UserRepo handles Database access
type UserRepository struct {
	serviceClient aztables.ServiceClient
}

// NewUserRepo creates and returns a new, unconnected UserRepo object
func NewUserRepo(cfg UserRepoConfig) services.UserRepo {
	cred, err := aztables.NewSharedKeyCredential(cfg.accountName, cfg.accountKey)
	handle(err)
	client, err := aztables.NewServiceClientWithSharedKey(cfg.serviceEndpointURL, cred, nil)
	handle(err)
	return &UserRepository{serviceClient: *client}
}

// CreateUser creates an aztable entity in the specified table name, creating the table if it doesn't exist
func (repo *UserRepository) CreateUser(tableName string, user models.User) error {

	//https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
	userEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: "Users",
			RowKey:       user.ID,
		},
		Properties: map[string]any{
			"Username": user.Name,
			"Email":    user.Email,
			"Role":     user.Role,
		},
	}

	//https://pkg.go.dev/encoding/json
	serializedEntity, err := json.Marshal(userEntity)
	handle(err)

	//TODO: Better handling?
	_, err = repo.serviceClient.CreateTable(context.Background(), tableName, nil)

	tableClient := repo.serviceClient.NewClient(tableName)

	_, err = tableClient.AddEntity(context.Background(), serializedEntity, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetUser retrieves and stores entity data in a User object
func (repo *UserRepository) GetUser(tableName string, id string) (models.User, error) {

	ctx := context.Background()
	pKey := "Users"
	tableClient := repo.serviceClient.NewClient(tableName)

	resp, err := tableClient.GetEntity(ctx, pKey, id, nil)
	if err != nil {
		return models.User{}, err
	}

	var myEntity aztables.EDMEntity
	err = json.Unmarshal(resp.Value, &myEntity)
	handle(err)

	user := models.User{
		ID:    myEntity.RowKey,
		Name:  myEntity.Properties["Username"].(string),
		Email: myEntity.Properties["Email"].(string),
		Role:  myEntity.Properties["Role"].(string),
	}

	return user, nil
}

// Generic error handler
func handle(err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
