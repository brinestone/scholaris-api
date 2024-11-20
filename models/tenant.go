package models

import (
	"database/sql"
	"time"
)

type Tenant struct {
	Name         string
	Id           uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Subscription uint64
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
