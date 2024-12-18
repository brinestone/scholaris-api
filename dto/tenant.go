package dto

import (
	"errors"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

type TenantNameAvailableResponse struct {
	Available bool `json:"available"`
}

type TenantNameAvailableRequest struct {
	Name string `query:"name"`
}

func (t TenantNameAvailableRequest) Validate() error {
	if len(t.Name) == 0 {
		return errors.New("invalid value for name")
	}
	return nil
}

type SubscriptionPlanBenefit struct {
	Name     string  `json:"name"`
	Details  *string `json:"details,omitempty" encore:"optional"`
	MinCount *int32  `json:"minCount,omitempty" encore:"optional"`
	MaxCount *int32  `json:"maxCount,omitempty" encore:"optional"`
}

type SubscriptionPlan struct {
	Id           uint64                    `json:"id"`
	Name         string                    `json:"name"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
	Price        *float64                  `json:"price,omitempty" encore:"optional"`
	Currency     *string                   `json:"currency,omitempty" encore:"optional"`
	Enabled      bool                      `json:"enabled"`
	BillingCycle uint                      `json:"billingCycle"`
	Benefits     []SubscriptionPlanBenefit `json:"benefits"`
}

type FindSubscriptionPlansResponse struct {
	Plans []SubscriptionPlan `json:"plans"`
}

type FindTenantsRequest struct {
	After uint64 `query:"after"`
	Size  uint   `query:"size"`
}

type FindTenantResponse struct {
	Tenants []TenantLookup `json:"tenants"`
}

type TenantLookup struct {
	Name             string    `json:"name"`
	Id               uint64    `json:"id"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	SubscriptionPlan string    `json:"subscriptionPlan"`
}

type NewSubscriptionPlan struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

type NewTenantRequest struct {
	Name         string `json:"name"`
	CaptchaToken string `json:"captchaToken"`
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

	if len(msgs) > 0 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}
	return nil
}
