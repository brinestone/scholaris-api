package dto

import (
	"errors"
	"strings"

	"encore.dev"
	"github.com/brinestone/scholaris/helpers"
)

const (
	ProvInternal = "internal"
	ProvClerk    = "clerk"
)

var (
	ClerkIssuerDomain     = "clerk.accounts.dev"
	ScholarisIssuerDomain = encore.Meta().APIBaseURL.Hostname()
	ValidIssuerDomains    = helpers.SliceOf(ClerkIssuerDomain, ScholarisIssuerDomain)
)

type AuthClaims struct {
	Email      string  `json:"email"`
	Avatar     *string `json:"avatar"`
	Provider   string  `json:"provider"`
	ExternalId string  `json:"externalId"`
	FullName   string  `json:"displayName"`
	Sub        uint64  `json:"sub"`
	Account    uint64  `json:"account"`
}

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
