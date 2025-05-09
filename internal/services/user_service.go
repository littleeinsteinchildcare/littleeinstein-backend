package services

import (
	"littleeinsteinchildcare/backend/internal/models"
)

const USERSTABLE = "UsersTable"

// UserRepo interface methods implemented in repositories package
type UserRepo interface {
	CreateUser(tableName string, user models.User) error
	GetUser(tableName string, id string) (models.User, error)
	GetAllUsers(tableName string) ([]models.User, error)
	DeleteUser(tableName string, id string) error
	UpdateUser(tableName string, user models.User) (models.User, error)
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

func (s *UserService) GetAllUsers() ([]models.User, error) {
	users, err := s.repo.GetAllUsers(USERSTABLE)
	if err != nil {
		return []models.User{}, err
	}
	return users, nil
}

// CreateUser returns an error on a failed UserRepo call
func (s *UserService) CreateUser(user models.User) error {
	err := s.repo.CreateUser(USERSTABLE, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) UpdateUser(user models.User) (models.User, error) {
	user, err := s.repo.UpdateUser(USERSTABLE, user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (s *UserService) DeleteUserByID(id string) error {
	err := s.repo.DeleteUser(USERSTABLE, id)
	if err != nil {
		return err
	}
	return nil
}
