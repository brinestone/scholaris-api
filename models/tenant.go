package models

import "time"

type Tenant struct {
	Name      string    `json:"name"`
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NewTenantRequest struct {
	Name string `json:"name" binding:"required"`
}
