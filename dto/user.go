package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

type LoginRequest struct {
	// The user's email address
	Email string `json:"email"`
	// The user's plaintext password
	Password string `json:"password" encore:"sensitive"`
	// The reCaptcha token for the site
	CaptchaToken string `json:"captchaToken"`
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

	if len(l.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token")
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
	// The user's first name
	FirstName string `json:"firstName"`
	// The user's last name (optional)
	LastName string `json:"lastName,omitempty"`
	// The user's email address
	Email string `json:"email"`
	// The user's date of birth (YYYY/MM/DD)
	Dob string `json:"dob"`
	// The user's plaintext password
	Password string `json:"password"`
	// Password verification
	ConfirmPassword string `json:"confirmPassword"`
	// The user's phone number in IE64 format
	Phone string `json:"phone,omitempty"`
	// The user's gender
	Gender Gender `json:"gender,omitempty"`
	// The captcha token for the request
	CaptchaToken string `json:"captchaToken"`
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

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token")
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
