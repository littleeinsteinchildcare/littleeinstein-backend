package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/config"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

const PartitionKey = "Users"

// UserRepo handles Database access
type UserRepository struct {
	serviceClient aztables.ServiceClient
}

// NewUserRepo creates and returns a new, unconnected UserRepo object
func NewUserRepo(cfg config.AzTableConfig) (services.UserRepo, error) {
	cred, err := aztables.NewSharedKeyCredential(cfg.AzureAccountName, cfg.AzureAccountKey)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.NewUserRepo: Failed to create credentials: %w", err)
	}
	client, err := aztables.NewServiceClientWithSharedKey(cfg.AzureContainerName, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.NewUserRepo: Failed to initialize service client: %w", err)
	}
	return &UserRepository{serviceClient: *client}, nil
}

// GetUser retrieves and stores entity data in a User object
func (repo *UserRepository) GetUser(tableName string, id string) (models.User, error) {

	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	resp, err := tableClient.GetEntity(ctx, PartitionKey, id, nil)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.GetUser: Failed to retrieve entity from %s: %w", tableName, err)
	}

	var myEntity aztables.EDMEntity
	err = json.Unmarshal(resp.Value, &myEntity)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.GetUser: Failed to deserialize entity: %w", err)
	}

	user := models.User{
		ID:    myEntity.RowKey,
		Name:  myEntity.Properties["Username"].(string),
		Email: myEntity.Properties["Email"].(string),
		Role:  myEntity.Properties["Role"].(string),
	}

	if entityImages, ok := myEntity.Properties["Images"]; ok {
		if imagesString, ok := entityImages.(string); ok {
			var images []string
			if err := json.Unmarshal([]byte(imagesString), &images); err != nil {
				return models.User{}, fmt.Errorf("UserRepository.GetUser: Failed to parse Image IDs")
			}
			user.Images = images
		}
	}
	return user, nil
}

func (repo *UserRepository) GetAllUsers(tableName string) ([]models.User, error) {
	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)
	filter := "PartitionKey eq 'Users'"
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
	}
	var users []models.User

	pager := tableClient.NewListEntitiesPager(options)
	pageCount := 0
	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return []models.User{}, fmt.Errorf("UserRepository.GetAllUsers: Failed to acquire next page: %w", err)
		}
		pageCount += 1

		for _, userEntity := range response.Entities {
			var myEntity aztables.EDMEntity
			err = json.Unmarshal(userEntity, &myEntity)
			if err != nil {
				return []models.User{}, fmt.Errorf("UserRepositroy.GetAllUser: Failed to Unmarshal entity: %w", err)
			}
			user := models.User{
				ID:    myEntity.RowKey,
				Name:  myEntity.Properties["Username"].(string),
				Email: myEntity.Properties["Email"].(string),
				Role:  myEntity.Properties["Role"].(string),
			}

			if entityImages, ok := myEntity.Properties["Images"]; ok {
				if imagesString, ok := entityImages.(string); ok {
					var images []string
					if err := json.Unmarshal([]byte(imagesString), &images); err != nil {
						return []models.User{}, fmt.Errorf("UserRepository.GetAllUsers: Failed to parse Image IDs")
					}
					user.Images = images
				}
			}

			users = append(users, user)
		}

	}
	return users, nil

}

// CreateUser creates an aztable entity in the specified table name, creating the table if it doesn't exist
func (repo *UserRepository) CreateUser(tableName string, user models.User) error {

	var imagesStr string
	if bytes, err := json.Marshal(user.Images); err == nil {
		imagesStr = string(bytes)
	} else {
		return err
	}

	//https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
	userEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: PartitionKey,
			RowKey:       user.ID,
		},
		Properties: map[string]any{
			"Username": user.Name,
			"Email":    user.Email,
			"Role":     user.Role,
			"Images":   imagesStr,
		},
	}

	//https://pkg.go.dev/encoding/json
	serializedEntity, err := json.Marshal(userEntity)
	if err != nil {
		return fmt.Errorf("UserRepository.CreateUser: Failed to serialize userEntity: %w", err)
	}

	//TODO: Better handling?
	_, err = repo.serviceClient.CreateTable(context.Background(), tableName, nil)
	tableClient := repo.serviceClient.NewClient(tableName)
	_, err = tableClient.AddEntity(context.Background(), serializedEntity, nil)
	if err != nil {
		return fmt.Errorf("UserRepository.CreateUser: Failed to add entity to table %s: %w", tableName, err)
	}
	return nil
}

func updateImages(newUserData models.User, user models.User) []string {
	imageSet := make(map[string]struct{})
	for _, id := range user.Images {
		imageSet[id] = struct{}{}
	}
	for _, id := range newUserData.Images {
		imageSet[id] = struct{}{}
	}
	uniqueList := make([]string, 0, len(imageSet))
	for id := range imageSet {
		uniqueList = append(uniqueList, id)
	}
	return uniqueList
}

func (repo *UserRepository) UpdateUser(tableName string, newUserData models.User) (models.User, error) {
	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	user, err := repo.GetUser(tableName, newUserData.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.UpdateUser: Failed to retrieve user ID %s from %s: %w", newUserData.ID, tableName, err)
	}

	if len(newUserData.Images) > len(user.Images) {
		newUserData.Images = updateImages(newUserData, user)
	}

	err = user.Update(newUserData)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.UpdateUser: Failed to update user ID %s's fields: %w", user.ID, err)
	}

	var imagesStr string
	if bytes, err := json.Marshal(user.Images); err == nil {
		imagesStr = string(bytes)
	} else {
		return models.User{}, err
	}

	userEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: PartitionKey,
			RowKey:       user.ID,
		},
		Properties: map[string]any{
			"Username": user.Name,
			"Email":    user.Email,
			"Role":     user.Role,
			"Images":   imagesStr,
		},
	}

	serializedEntity, err := json.Marshal(userEntity)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.UpdateUser: Failed to serialize user data: %w", err)
	}
	_, err = tableClient.UpdateEntity(ctx, serializedEntity, nil)
	if err != nil {
		return models.User{}, fmt.Errorf("UserRepository.UpdateEntity: Failed to update entity in %s: %w", tableName, err)
	}

	return user, nil
}

func (repo *UserRepository) DeleteUser(tableName string, id string) error {
	ctx := context.Background()
	tableClient := repo.serviceClient.NewClient(tableName)

	options := &aztables.DeleteEntityOptions{
		IfMatch: to.Ptr(azcore.ETagAny),
	}
	_, err := tableClient.DeleteEntity(ctx, PartitionKey, id, options)
	if err != nil {
		return fmt.Errorf("UserRepository.DeleteUser: Failed to delete entity with ID %s from %s: %w", id, tableName, err)
	}

	return nil
}
