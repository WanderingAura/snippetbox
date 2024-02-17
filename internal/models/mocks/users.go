package mocks

import (
	"snippetbox.volcanoeyes.net/internal/models"
)

var mockUser = &models.User{
	ID:    1,
	Name:  "Alice",
	Email: "alice@example.com",
}

type UserModel struct{}

func (m *UserModel) Get(id int) (*models.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "bigcactus" {
		return 1, nil
	}
	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) UpdatePassword(id int, currentPassword, newPassword string) error {
	return nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}
