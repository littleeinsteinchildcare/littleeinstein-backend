package services

import (
	"littleeinsteinchildcare/backend/internal/models"
	"log"
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
	log.Printf("DEBUG: Calling GetUserByID In UserService")
	user, err := s.repo.GetUser(USERSTABLE, id)

	if err != nil {
		log.Printf("DEBUG: GET USER FROM REPO FAILED")
		return models.User{}, err
	}
	return user, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	log.Printf("DEBUG: USER SERVICE - GET ALL USERS: CALLED")
	users, err := s.repo.GetAllUsers(USERSTABLE)
	if err != nil {
		log.Printf("DEBUG: USER SERVICE - GET ALL USERS FROM REPO: FAILED")
		return []models.User{}, err
	}
	log.Printf("DEBUG: USER SERVICE - GET ALL USERS: SUCCESS")
	return users, nil
}

// CreateUser returns an error on a failed UserRepo call
func (s *UserService) CreateUser(user models.User) error {
	log.Printf("DEBUG: CREATE USER CALLED IN USER SERVICE")
	err := s.repo.CreateUser(USERSTABLE, user)
	if err != nil {
		log.Printf("DEBUG: CREATE USER FAILED IN USER SERVICE")
		return err
	}
	log.Printf("DEBUG: USER SERVICE - CREATE USER: SUCCESS")
	return nil
}

// Update User and Process Errors from User Repo
func (s *UserService) UpdateUser(user models.User) (models.User, error) {
	log.Printf("DEBUG: UPDATE USER CALLED IN USER SERVICE")
	user, err := s.repo.UpdateUser(USERSTABLE, user)
	if err != nil {
		log.Printf("DEBUG: UPDATE USER FAILED IN USER SERVICE")
		return user, err
	}
	return user, nil
}

// Delete User from Users Table and all relevant Events and Invitations
func (s *UserService) DeleteUserByID(id string) error {
	log.Printf("DEBUG: USER SERVICE - DELETE USER BY ID: CALLED")
	err1 := s.repo.DeleteUser(USERSTABLE, id)
	if err1 != nil {
		log.Printf("DEBUG: USER SERVICE - DELETE USER FROM REPO: RETURNED WITH ERROR %v", err1)
		return err1
	}
	log.Printf("DEBUG: USER SERVICE - DELETE USER FROM REPO: RETURNED SUCCESS")

	err2 := s.eventRepo.DeleteEventByUserID(EVENTSTABLE, id)
	if err2 != nil {
		log.Printf("DEBUG: USER SERVICE - DELETE EVENTS BY USER ID: RETURNED WITH ERROR %v", err1)
		return err2
	}
	log.Printf("DEBUG: USER SERVICE - EVENT REPO DELETE EVENT BY USER ID: RETURNED SUCCESS")

	err3 := s.eventRepo.RemoveInvitee(EVENTSTABLE, id)
	if err3 != nil {
		log.Printf("DEBUG: USER SERVICE - EVENT REPO REMOVE INVITEES: RETURNED WITH ERROR %v", err1)
		return err3
	}
	log.Printf("DEBUG: USER SERVICE - EVENT REPO REMOVE INVITEE: RETURNED SUCCESS")

	err4 := s.blobRepo.DeleteAllImages(id)
	if err4 != nil {
		log.Printf("DEBUG: USER SERVICE - BLOB REPO DELETE ALL IMAGES: RETURNED WITH ERROR %v", err1)
		return err4
	}
	log.Printf("DEBUG: USER SERVICE - BLOB REPO DELETE ALL IMAGES: RETURNED SUCCESS")

	return nil
}
