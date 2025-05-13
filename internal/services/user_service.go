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
	repo      UserRepo
	eventRepo EventRepo
}

// NewUserService constructs and returns a UserService object
func NewUserService(r UserRepo, e EventRepo) *UserService {
	return &UserService{repo: r, eventRepo: e}
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

// Update User and Process Errors from User Repo
func (s *UserService) UpdateUser(user models.User) (models.User, error) {
	user, err := s.repo.UpdateUser(USERSTABLE, user)
	if err != nil {
		return user, err
	}
	return user, nil
}

// Delete User from Users Table and all relevant Events and Invitations
func (s *UserService) DeleteUserByID(id string) error {
	err1 := s.repo.DeleteUser(USERSTABLE, id)
	if err1 != nil {
		return err1
	}
	err2 := s.eventRepo.DeleteEventByUserID(EVENTSTABLE, id)
	if err2 != nil {
		return err2
	}
	err3 := s.eventRepo.RemoveInvitee(EVENTSTABLE, id)
	if err3 != nil {
		return err3
	}

	return nil
}
