package models

import (
	"database/sql"
	"time"
)

type Tenant struct {
	Name         string    `json:"name"`
	Id           uint64    `json:"id,omitempty" encore:"optional"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Subscription uint64    `json:"subscription"`
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
