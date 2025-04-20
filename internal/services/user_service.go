package services

import (
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
)

const TABLENAME = "UsersTable"

// UserRepo interface methods implemented in repositories package
type UserRepo interface {
	CreateUser(tableName string, user models.User) error
	GetUser(tableName string, id string) (models.User, error)
}

// UserService contains and handles a specific UserRepository object
type UserService struct {
	repo UserRepo
}

// NewUserService constructs and returns a UserService object
func NewUserService(r UserRepo) *UserService {
	return &UserService{repo: r}
}

// GetUserByID handles calling the UserRepository GetUser function and returns the result of a query by the UserRepository
func (s *UserService) GetUserByID(id string) (models.User, error) {
	user, err := s.repo.GetUser(TABLENAME, id)
	if err != nil {
		return models.User{}, err
	}
	fmt.Printf("User: %v", user)
	return user, nil
}

// CreateUser returns an error on a failed UserRepo call
func (s *UserService) CreateUser(user models.User) error {
	err := s.repo.CreateUser(TABLENAME, user)
	if err != nil {
		return err
	}
	return nil
}
