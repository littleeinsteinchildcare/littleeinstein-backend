package services

import (
	"littleeinsteinchildcare/backend/internal/models"
)

const USERSTABLE = "UsersTable"

// UserRepo interface methods implemented in repositories package
type UserRepo interface {
	CreateUser(tableName string, user models.User) error
	GetUser(tableName string, id string) (models.User, error)
	DeleteUser(tableName string, id string) (bool, error)
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
	user, err := s.repo.GetUser(USERSTABLE, id)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// CreateUser returns an error on a failed UserRepo call
func (s *UserService) CreateUser(user models.User) error {
	err := s.repo.CreateUser(USERSTABLE, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) DeleteUserByID(id string) (bool, error) {
	success, err := s.repo.DeleteUser(USERSTABLE, id)
	if err != nil {
		return success, err
	}
	return success, nil
}
