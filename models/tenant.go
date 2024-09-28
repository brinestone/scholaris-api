package models

import "time"

type Tenant struct {
	Name      string    `json:"name"`
	Id        uint64    `json:"id,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
