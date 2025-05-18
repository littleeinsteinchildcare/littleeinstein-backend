package models

import "errors"

type User struct {
	ID       string
	Name     string
	Email    string
	Role     string
	ImageIDs []string
}

func NewUser(id string, name string, email string, role string, imageIDs []string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Role:     role,
		ImageIDs: imageIDs,
	}
}

func (userModel *User) Update(newUserData User) error {

	if newUserData.ID != userModel.ID {
		return errors.New("Invalid ID when trying to update fields in User")
	}
	if newUserData.Name != "" {
		userModel.Name = newUserData.Name
	}
	if newUserData.Email != "" {
		userModel.Email = newUserData.Email
	}
	if newUserData.Role != "" {
		userModel.Role = newUserData.Role
	}
	if len(newUserData.ImageIDs) > 0 {
		userModel.ImageIDs = newUserData.ImageIDs
	}
	return nil

}
