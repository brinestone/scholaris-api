package dto

import (
	"errors"
	"strings"
	"time"
)

type NewFormInput struct {
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	BackgroundColor *string `json:"backgroundColor,omitempty"`
	BackgroundImage *string `json:"backgroundImage,omitempty"`
	Image           *string `json:"image,omitempty"`
	MultiResponse   bool    `json:"multiResponse"`
	Resubmission    bool    `json:"resubmission"`
	CaptchaToken    string  `header:"x-ver-token"`
	Owner           uint64  `header:"x-owner"`
	OwnerType       string  `header:"x-owner-type"`
}

func (n NewFormInput) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewFormInput) Validate() error {
	var msgs = make([]string, 0)

	if len(n.Title) == 0 {
		msgs = append(msgs, "The title field is required")
	}

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token failed")
	}

	if len(n.OwnerType) == 0 {
		msgs = append(msgs, "x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(n.OwnerType); !ok {
		msgs = append(msgs, "Invalid owner type")
	}

	if n.Owner == 0 {
		msgs = append(msgs, "The x-owner field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type FormConfig struct {
	Id              uint64    `json:"id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	BackgroundColor *string   `json:"backgroundColor,omitempty"`
	Status          string    `json:"status"`
	BackgroundImage *string   `json:"backgroundImage,omitempty"`
	Image           *string   `json:"image,omitempty"`
	MultiResponse   bool      `json:"multiResponse"`
	Resubmission    bool      `json:"resubmission"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdateAt        time.Time `json:"updatedAt"`
}
