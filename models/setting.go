package models

import (
	"database/sql"
	"time"
)

type SettingOption struct {
	Id      uint64  `json:"id"`
	Label   string  `json:"label"`
	Value   *string `json:"value"`
	Setting uint64  `json:"setting"`
}

type SettingValue struct {
	Id      uint64     `json:"id"`
	Index   uint       `json:"index"`
	Setting uint64     `json:"setting"`
	SetAt   *time.Time `json:"setAt"`
	SetBy   uint64     `json:"setBy"`
}

type Setting struct {
	Id              uint64
	Label           string
	Description     sql.NullString
	Key             string
	MultiValues     bool
	SystemGenerated bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Parent          sql.NullInt64
	Owner           uint64
	OwnerType       string
	Overridable     bool
	CreatedBy       uint64
	Options         []SettingOption
	Values          []SettingValue
}
