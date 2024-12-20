package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/brinestone/scholaris/helpers"
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

func (d *DateOnly) Scan(value any) (err error) {
	if value != nil {
		d.Time, d.Valid = value.(time.Time)
	}
	return
}

func (d *DateOnly) Value() (ans any, err error) {
	if !d.Valid {
		return
	}
	ans = d.Time.Format(time.DateOnly)
	return
}

func (d *DateOnly) UnmarshalJSON(b []byte) (err error) {
	if len(b) <= 2 || string(b) == "null" || string(b) == "NULL" {
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

func (u *User) GetAvatar() (ans *string) {
	avatars := helpers.SliceMap(u.ProvidedAccounts, func(a UserAccount) *string {
		return a.ImageUrl
	})

	ans = helpers.Coalesce(avatars...)
	return
}

func (u *User) FullName() (ans *string) {
	tmp := u.ProvidedAccounts[0].FullName()
	return &tmp
}

func (u *User) GetEmail() (ans *string) {
	if u.PrimaryEmail.Valid {
		email, found := helpers.Find(u.Emails, func(a UserEmailAddress) bool {
			return a.IsPrimary
		})
		if found {
			ans = &email.Email
		}
	} else if len(u.Emails) > 0 {
		ans = &u.Emails[0].Email
	}
	return
}

// Gets any phone number assigned to this user.
// The user's primary phone number takes higher preference over others.
func (u *User) GetPhoneNumber() (ans *string) {
	if u.PrimaryPhone.Valid {
		phone, found := helpers.Find(u.PhoneNumbers, func(a UserPhoneNumber) bool {
			return a.IsPrimary
		})
		if found {
			ans = &phone.Phone
		}
	} else if len(u.PhoneNumbers) > 0 {
		ans = &u.PhoneNumbers[0].Phone
	}
	return
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
