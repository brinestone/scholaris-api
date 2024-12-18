package dto

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

type CreateTenantInviteRequest struct {
	Email           string  `json:"email"`
	Phone           *string `json:"phone,omitempty" encore:"optional"`
	Names           string  `json:"displayName"`
	SuccessRedirect string  `json:"redirecUrl"`
	OnboardRedirect string  `json:"onboardRedirect"`
	ErrorRedirect   string  `json:"errorRedirect"`
	CaptchaToken    string  `json:"captcha"`
}

func (c CreateTenantInviteRequest) GetCaptchaToken() string {
	return c.CaptchaToken
}

func (c CreateTenantInviteRequest) Validate() (err error) {
	msgs := make([]string, 0)

	if len(c.Names) == 0 {
		msgs = append(msgs, "The displayName field is required")
	}

	if len(c.CaptchaToken) == 0 {
		msgs = append(msgs, "The captcha field is required")
	}

	if len(c.ErrorRedirect) == 0 {
		msgs = append(msgs, "The errorRedirect field is required")
	}

	if len(c.SuccessRedirect) == 0 {
		msgs = append(msgs, "The redirectUrl field is required")
	}

	if len(c.OnboardRedirect) == 0 {
		msgs = append(msgs, "The onboardRedirect field is required")
	}

	if len(c.Email) == 0 {
		msgs = append(msgs, "The email field is required")
	} else {
		emailValid := regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`).MatchString(c.Email)
		if !emailValid {
			msgs = append(msgs, "Invalid email address")
		}
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}
	return
}

type TenantMembership struct {
	Id              *uint64            `json:"id,omitempty" encore:"optional"`
	Invite          uint64             `json:"invite"`
	User            uint64             `json:"user"`
	Tenant          uint64             `json:"tenant"`
	DisplayName     string             `json:"displayName"`
	Email           string             `json:"email"`
	InviteStatus    string             `json:"inviteStatus"`
	Role            string             `json:"role"`
	Avatar          *string            `json:"avatar,omitempty" encore:"optional"`
	Phone           *string            `json:"phone,omitempty" encore:"optional"`
	Prefs           *map[string]string `json:"prefs,omitempty" encore:"optional"`
	InvitedAt       time.Time          `json:"invitedAt"`
	InviteExpiresAt *time.Time         `json:"inviteExpiresAt,omitempty" encore:"optional"`
	JoinedAt        *time.Time         `json:"joinedAt,omitempty" encore:"optional"`
	UpdatedAt       *time.Time         `json:"updatedAt,omitempty" encore:"optional"`
}

type FindTenantMembersResponse struct {
	// The tenant memberships
	Members []TenantMembership `json:"members"`
}

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

type NewTenantResponse struct {
	Id uint64 `json:"id"`
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
