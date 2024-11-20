package dto

import (
	"errors"
	"strings"
)

type DeleteAccountRequest struct {
	CaptchaToken string `header:"x-captcha" encore:"sensitive"`
	Password     string `json:"password" encore:"sensitive"`
}

func (d DeleteAccountRequest) GetCaptchaToken() string {
	return d.CaptchaToken
}

func (d DeleteAccountRequest) Validate() (err error) {
	msgs := make([]string, 0)

	if len(d.CaptchaToken) == 0 {
		msgs = append(msgs, "The x-captcha header is required")
	}

	if len(d.Password) == 0 {
		msgs = append(msgs, "The password field is required")
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}

	return
}
