package main

import (
	"time"

	"github.com/adamararcane/d2optifarm/backend/internal/database"
)

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

func databaseUserToUser(user database.User) (User, error) {
	createdAt, err := time.Parse(time.RFC3339, user.CreatedAt)
	if err != nil {
		return User{}, err
	}

	updatedAt, err := time.Parse(time.RFC3339, user.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:        user.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}, nil
}
