package dto

import (
	"strings"
	"time"

	"encore.dev/beta/errs"
)

type FindTenantsRequest struct {
	After uint64 `query:"after"`
	Size  uint   `query:"size"`
	// Retrieve only the tenants whereby the user is a member.
	SubscribedOnly bool `json:"subscribedOnly"`
}
type TenantLookup struct {
	Name         string    `json:"name"`
	Id           uint64    `json:"id,omitempty" encore:"optional"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Subscription uint64    `json:"-"`
}

type NewSubscriptionPlan struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

type NewTenantRequest struct {
	Name               string `json:"name"`
	SubscriptionPlanId uint64 `json:"subscriptionPlan,omitempty" encore:"optional"`
	CaptchaToken       string `json:"captchaToken"`
}

func (n NewTenantRequest) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewTenantRequest) Validate() error {
	var msgs = make([]string, 0)

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "The captchaToken field is required")
	}

	if len(n.Name) == 0 {
		msgs = append(msgs, "The name field is required")
	}

	if n.SubscriptionPlanId == 0 {
		msgs = append(msgs, "The subscriptionPlan field is required")
	}

	if len(msgs) > 0 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}
	return nil
}
