package models

import "errors"

type User struct {
	ID    string
	Name  string
	Email string
	Role  string
}

func NewUser(id string, name string, email string, role string) *User {
	return &User{
		ID:    id,
		Name:  name,
		Email: email,
		Role:  role,
	}
}

func (userModel *User) UpdateFields(newUserData User) error {

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
	return nil

}
