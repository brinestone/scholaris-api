package models

import "time"

type Institution struct {
	Name        string     `json:"name"`
	Description NullString `json:"description,omitempty"`
	Logo        NullString `json:"logo,omitempty"`
	Visible     bool       `json:"visible"`
	Slug        string     `json:"slug"`
	Id          uint64     `json:"id,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
