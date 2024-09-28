package dto

import (
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

type NewUserRequest struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	Dob             string `json:"dob"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	Phone           string `json:"phone"`
	Gender          string `json:"gender"`
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

	if len(n.Gender) == 0 {
		msgs = append(msgs, "Invalid gender")
	}

	if len(msgs) > 0 {
		ans = &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}

	return ans
}

var emailRegex = regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`)
