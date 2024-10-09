package dto

import (
	"strings"
	"time"

	"encore.dev/beta/errs"
	"github.com/brinestone/scholaris/models"
)

type InstitutionLookup struct {
	Name        string            `json:"name"`
	Description models.NullString `json:"description,omitempty"`
	Logo        models.NullString `json:"logo,omitempty"`
	Visible     bool              `json:"visible"`
	Slug        string            `json:"slug"`
	Id          uint64            `json:"id,omitempty"`
	TenantId    uint64            `json:"-"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	IsMember    bool              `json:"isMember"`
}

type NewInstitutionRequest struct {
	// The institution's name
	Name string `json:"name"`
	// The institution's description (optional)
	Description string `json:"description,omitempty"`
	// The institution's logo (optional)
	Logo string `json:"logo,omitempty"`
	// The institution's slug
	Slug string `json:"slug"`
	// The institution's tenant ID
	TenantId uint64 `json:"tenantId"`
	// The request's captcha token
	Captcha string `json:"captcha"`
	// The timestamp of the request
	Timestamp time.Time `header:"x-timestamp"`
}

func (n NewInstitutionRequest) GetCaptchaToken() string {
	return n.Captcha
}

func (n NewInstitutionRequest) Validate() error {
	msgs := make([]string, 0)

	if n.TenantId == 0 {
		msgs = append(msgs, "The tenantId field is required")
	}

	if len(n.Name) == 0 {
		msgs = append(msgs, "The name field is required")
	}

	if len(n.Slug) == 0 {
		msgs = append(msgs, "The slug field is required")
	}

	if len(msgs) > 0 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}

	return nil
}
