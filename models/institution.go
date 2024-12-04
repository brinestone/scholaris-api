package models

import (
	"database/sql"
	"time"
)

type AcademicTerm struct {
	Id          uint64
	Year        uint64
	Institution uint64
	StartDate   time.Time
	CreatedAt   time.Time
	Duration    time.Duration
	EndDate     time.Time
	Label       string
	UpdatedAt   time.Time
}

type AcademicYear struct {
	Id          uint64
	Institution uint64
	StartDate   time.Time
	Duration    time.Duration
	EndDate     time.Time
	Label       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Terms       []AcademicTerm
}

type InstitutionStatistics struct {
	TotalInstitutions uint64
	TotalVerified     uint64
	TotalUnverified   uint64
}

type Institution struct {
	Name        string
	Description sql.NullString
	Logo        sql.NullString
	Visible     bool
	Slug        string
	Id          uint64
	TenantId    uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Verified    bool
	CurrentYear sql.NullInt64
	CurrentTerm sql.NullInt64
}
