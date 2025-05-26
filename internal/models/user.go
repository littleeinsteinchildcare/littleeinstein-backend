package models

import (
	"errors"
)

type User struct {
	ID     string
	Name   string
	Email  string
	Role   string
	Images []string
}

func NewUser(id string, name string, email string, role string, images []string) *User {
	return &User{
		ID:     id,
		Name:   name,
		Email:  email,
		Role:   role,
		Images: images,
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

	if newUserData.Images != nil {
		if len(newUserData.Images) >= 3 {
			return errors.New("User Max number of Images exceeded")
		} else {
			userModel.Images = newUserData.Images
			cleaned := []string{}
			for _, img := range newUserData.Images {
				if img != "" {
					cleaned = append(cleaned, img)
				}
			}
			userModel.Images = cleaned

		}
	} else {
		cleaned := []string{}
		for _, img := range newUserData.Images {
			if img != "" {
				cleaned = append(cleaned, img)
			}
		}
		userModel.Images = cleaned
	}
	return nil
}

func (userModel *User) UpdateImages(newUserData User) error {
	if len(newUserData.Images) >= 3 {
		return errors.New("User Max Number of Images exceeded")
	} else {
		userModel.Images = newUserData.Images
	}
	return nil
}
