package models

import (
	"database/sql"
	"time"
)

type Institution struct {
	Name        string
	Description NullString
	Logo        NullString
	Visible     bool
	Slug        string
	Id          uint64
	TenantId    uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SettingValue struct {
	Id      uint64    `json:"id"`
	Setting uint64    `json:"setting"`
	Value   string    `json:"value"`
	SetBy   uint64    `json:"setBy"`
	SetAt   time.Time `json:"setAt"`
}

type InstitutionSetting struct {
	Id              uint64
	Institution     uint64
	SystemGenerated bool
	Key             string
	Label           string
	Description     sql.NullString
	MultiValue      bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	IsRequired      bool
	UpdatedBy       uint64
	Parent          sql.NullInt64
	ParentType      sql.NullString
	Values          []SettingValue
}
