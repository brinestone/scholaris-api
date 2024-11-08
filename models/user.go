package models

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	Id           uint64     `json:"id,omitempty" encore:"optional"`
	FirstName    string     `json:"firstName"`
	LastName     NullString `json:"lastName,omitempty" encore:"optional"`
	Email        string     `json:"email"`
	Dob          time.Time  `json:"dob"`
	PasswordHash string     `json:"-" encore:"sensitive"`
	Phone        string     `json:"phone"`
	CreatedAt    time.Time  `json:"created_at,omitempty" encore:"optional"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty" encore:"optional"`
	Gender       string     `json:"gender"`
	Avatar       NullString `json:"avatar,omitempty" encore:"optional"`
}

func (u User) FullName() string {
	return strings.Trim(fmt.Sprintf("%s %s", u.FirstName, u.LastName.String), "\t\n")
}
