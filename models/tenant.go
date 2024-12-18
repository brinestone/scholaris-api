package models

import (
	"database/sql"
	"time"
)

type TenantMembershipInvitation struct {
	Id                                                                           uint64
	User, Tenant                                                                 sql.NullInt64
	TenantName, Email, Role, Status                                              string
	Phone, RedirectUrl, ErrorRedirect, OnboardRedirect, Avatar, Url, DisplayName sql.NullString
	CreatedAt, UpdatedAt                                                         time.Time
	ExpiresAt                                                                    DateOnly
}

type TenantMembership struct {
	Id                                     sql.NullInt64
	Invite, User, Tenant                   uint64
	DisplayName, Email, InviteStatus, Role string
	Avatar, Phone                          sql.NullString
	Prefs                                  *map[string]string
	InvitedAt                              time.Time
	InviteExpiresAt                        *DateOnly
	CreatedAt, UpdatedAt                   sql.NullTime
}

type Tenant struct {
	Name             string
	Id               uint64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Subscription     uint64
	SubscriptionName string
}

type SubscriptionPlanBenefit struct {
	Name     string
	Details  *string
	MinCount *int32
	MaxCount *int32
}

type SubscriptionPlan struct {
	Id           uint64
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Price        sql.NullFloat64
	Currency     sql.NullString
	Enabled      bool
	BillingCycle uint
	Benefits     []SubscriptionPlanBenefit
}
