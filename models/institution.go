package models

import (
	"database/sql"
	"time"
)

type Institution struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description,omitempty" encore:"optional"`
	Logo        sql.NullString `json:"logo,omitempty" encore:"optional"`
	Visible     bool           `json:"visible"`
	Slug        string         `json:"slug"`
	Id          uint64         `json:"id,omitempty" encore:"optional"`
	TenantId    uint64         `json:"tenant"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}
