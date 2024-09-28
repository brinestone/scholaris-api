package models

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	Id           int64      `json:"id,omitempty"`
	FirstName    string     `json:"firstName"`
	LastName     NullString `json:"lastName,omitempty"`
	Email        string     `json:"email"`
	Dob          time.Time  `json:"dob"`
	PasswordHash string     `json:"-" encore:"sensitive"`
	Phone        string     `json:"phone"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
	Gender       string     `json:"gender"`
	Avatar       NullString `json:"avatar,omitempty"`
}

func (u User) FullName() string {
	return strings.Trim(fmt.Sprintf("%s %s", u.FirstName, u.LastName.String), "\t\n")
}
