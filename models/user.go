package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type User struct {
	Id           uint64
	FirstName    string
	LastName     sql.NullString
	Email        string
	Dob          time.Time
	PasswordHash string `encore:"sensitive"`
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Gender       string
	Avatar       sql.NullString
}

func (u User) FullName() string {
	return strings.Trim(fmt.Sprintf("%s %s", u.FirstName, u.LastName.String), "\t\n")
}
