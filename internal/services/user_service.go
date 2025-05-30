package services

import (
	"log"
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
	blobRepo  BlobRepo
}

// NewUserService constructs and returns a UserService object
func NewUserService(r UserRepo, e EventRepo, b BlobRepo) *UserService {
	return &UserService{repo: r, eventRepo: e, blobRepo: b}
}

// GetUserByID handles calling the UserRepository GetUser function and returns the result of a query by the UserRepository
func (s *UserService) GetUserByID(id string) (models.User, error) {
	log.Printf("DEBUG: UserService.GetUserByID called with id: '%s'", id)
	user, err := s.repo.GetUser(USERSTABLE, id)
	if err != nil {
		log.Printf("DEBUG: UserService.GetUserByID repo.GetUser failed: %v", err)
		return models.User{}, err
	}
	log.Printf("DEBUG: UserService.GetUserByID success: %+v", user)
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
	err4 := s.blobRepo.DeleteAllImages(id)
	if err4 != nil {
		return err4
	}

	// Return the deleted user data as a model
	// Grab the blob storage IDs from that model
	// Call the blob service and delete those by ID
	// Delete the entire blob as a user

	return nil
}