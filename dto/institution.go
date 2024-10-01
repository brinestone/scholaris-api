package dto

import (
	"strings"

	"encore.dev/beta/errs"
)

type NewInstitutionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Logo        string `json:"logo,omitempty"`
	Slug        string `json:"slug"`
	TenantId    uint64 `json:"tenantId"`
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
