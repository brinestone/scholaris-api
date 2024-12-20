package models

import (
	"database/sql"
	"time"
)

type ShortLink struct {
	Key                  string
	ClickCount           uint
	MaxClicks            sql.NullInt32
	ErrorRedirect        sql.NullString
	SuccessRedirect      string
	CreatedAt, UpdatedAt time.Time
	ExpiresAt            DateOnly
	IsUsable             bool
}
