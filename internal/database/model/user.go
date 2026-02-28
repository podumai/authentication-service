package model

import (
	"time"
)

type User struct {
	ID           int       `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"pass_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Permissions  []string  `json:"perms"`
	Status       string    `json:"status"`
}
