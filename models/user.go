package models

import "time"

type User struct {
	Id           int64     `json:"id,omitempty"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Dob          time.Time `json:"dob"`
	PasswordHash string    `encore:"sensitive"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Gender       string    `json:"gender"`
	Avatar       string    `json:"avatar,omitempty"`
}
