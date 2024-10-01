package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l LoginRequest) Validate() error {
	var err error
	var msgs = make([]string, 0)

	if len(l.Email) == 0 {
		msgs = append(msgs, "The email field is required")
	} else {
		if emailValid := emailRegex.MatchString(l.Email); !emailValid {
			msgs = append(msgs, "Invalid email address")
		}
	}

	if len(l.Password) == 0 {
		msgs = append(msgs, "The password field is required")
	}

	if len(msgs) > 0 {
		err = &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}

	return err
}

type UserLookupByEmailRequest struct {
	Email string `query:"email"`
}

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

type NewUserRequest struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName,omitempty"`
	Email           string `json:"email"`
	Dob             string `json:"dob"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Phone           string `json:"phone,omitempty"`
	Gender          Gender `json:"gender,omitempty"`
}

func (g Gender) Validate() error {
	if g != Male && g != Female {
		return fmt.Errorf("invalid gender value. Expected \"%s\" or \"%s\". Got: \"%s\"", Male, Female, g)
	}
	return nil
}

func (n NewUserRequest) Validate() error {
	var ans error
	msgs := make([]string, 0)

	if len(n.FirstName) == 0 {
		msgs = append(msgs, "The firstName field is required")
	}
	if len(n.Email) == 0 {
		msgs = append(msgs, "The email field is required")
	} else {
		if emailValid := emailRegex.MatchString(n.Email); !emailValid {
			msgs = append(msgs, "Invalid email address")
		}
	}

	if len(n.Dob) == 0 {
		msgs = append(msgs, "The dob field is required")
	} else {
		_, err := time.Parse("2006/2/1", n.Dob)
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}

	if len(n.Password) == 0 {
		msgs = append(msgs, "The password field is required")
	}

	if len(n.ConfirmPassword) == 0 {
		msgs = append(msgs, "The confirmPassword field is required")
	}

	if n.ConfirmPassword != n.Password {
		msgs = append(msgs, "Passwords do not match")
	}

	if err := n.Gender.Validate(); err != nil {
		msgs = append(msgs, err.Error())
	}

	if len(n.Phone) > 0 && !regexp.MustCompile(`\+(9[976]\d|8[987530]\d|6[987]\d|5[90]\d|42\d|3[875]\d|2[98654321]\d|9[8543210]|8[6421]|6[6543210]|5[87654321]|4[987654310]|3[9643210]|2[70]|7|1)\d{1,14}$`).MatchString(n.Phone) {
		msgs = append(msgs, "Invalid phone number. Phone numbers must be in international format")
	}

	if len(msgs) > 0 {
		ans = &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}

	return ans
}

var emailRegex = regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,}$`)
