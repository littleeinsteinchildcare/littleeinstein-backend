package models

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
