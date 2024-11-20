package dto

import (
	"errors"
	"strings"
)

type NewEnrollmentFormRequest struct {
	Institution  uint64 `header:"x-owner"`
	CaptchaToken string `json:"captcha"`
}

func (n NewEnrollmentFormRequest) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewEnrollmentFormRequest) GetOwner() uint64 {
	return n.Institution
}

func (n NewEnrollmentFormRequest) GetOwnerType() string {
	return string(PTInstitution)
}

func (n NewEnrollmentFormRequest) Validate() (err error) {
	msgs := make([]string, 0)
	if n.Institution == 0 {
		msgs = append(msgs, "The x-owner header is required")
	}

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token")
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}
	return
}

type NewEnrollmentRequest struct {
	Destination             uint64 `header:"x-owner"`
	ServiceTransactionToken string `header:"x-service-transaction"`
	Level                   uint64 `json:"level"`
	CaptchaToken            string `json:"captcha"`
}

func (n NewEnrollmentRequest) GetLevelRef() uint64 {
	return n.Level
}

func (n NewEnrollmentRequest) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewEnrollmentRequest) GetOwner() uint64 {
	return n.Destination
}

func (n NewEnrollmentRequest) GetOwnerType() string {
	return string(PTInstitution)
}

func (ne NewEnrollmentRequest) Validate() error {
	var msgs = make([]string, 0)

	if ne.Level == 0 {
		msgs = append(msgs, "Invalid level value")
	}

	if len(ne.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token")
	}

	if ne.Destination == 0 {
		msgs = append(msgs, "Invalid institution ID")
	}

	if len(ne.ServiceTransactionToken) == 0 {
		msgs = append(msgs, "Invalid Service Transaction Verification Token")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}
