package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type UserEmailAddress struct {
	Id         uint64
	Email      string
	Account    uint64
	ExternalId string
	IsPrimary  bool
	Verified   bool
}

type UserPhoneNumber struct {
	Id         uint64
	Phone      string
	Account    uint64
	ExternalId string
	IsPrimary  bool
	Verified   bool
}

type DateOnly struct {
	Valid bool
	time.Time
}

func (d *DateOnly) UnmarshalJSON(b []byte) (err error) {
	if len(b) <= 2 || string(b) == "null" {
		d.Valid = false
		return
	}

	actualValue := b[1 : len(b)-1]
	var date time.Time
	if !strings.Contains(string(actualValue), "T") {
		date, err = time.Parse(time.DateOnly, string(actualValue))
	} else {
		date, err = time.Parse(time.RFC3339, string(actualValue))
	}
	if err != nil {
		return
	}
	d.Valid = true
	d.Time = date
	return
}

type UserAccount struct {
	Id                  uint64
	ExternalId          string
	ImageUrl            *string
	User                uint64
	FirstName           *string
	LastName            *string
	Provider            string
	ProviderProfileData *string
	Gender              *string
	Dob                 *DateOnly
}
type User struct {
	Id               uint64
	Banned           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Locked           bool
	PrimaryEmail     sql.NullInt64
	PrimaryPhone     sql.NullInt64
	Emails           []UserEmailAddress
	ProvidedAccounts []UserAccount
	PhoneNumbers     []UserPhoneNumber
}

func (u UserAccount) FullName() string {
	lastName := ""
	if u.LastName != nil {
		lastName = *u.LastName
	}

	firstName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	return strings.Trim(fmt.Sprintf("%s %s", firstName, lastName), "\t\n")
}
