package models

import "time"

type Tenant struct {
	Name         string    `json:"name"`
	Id           uint64    `json:"id,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Subscription uint64    `json:"subscription"`
}

type SubscriptionPlanBenefit struct {
	Name     string     `json:"name"`
	Details  string     `json:"details"`
	MinCount *NullInt32 `json:"minCount"`
	MaxCount *NullInt32 `json:"maxCount"`
}

type SubscriptionPlan struct {
	Id           uint64                     `json:"id,omitempty"`
	Name         string                     `json:"name"`
	CreatedAt    time.Time                  `json:"createdAt"`
	UpdatedAt    time.Time                  `json:"updatedAt"`
	Price        float64                    `json:"price"`
	Currency     string                     `json:"currency"`
	Enabled      bool                       `json:"enabled"`
	BillingCycle uint                       `json:"billingCycle"`
	Benefits     *[]SubscriptionPlanBenefit `json:"benefits"`
}
