package models

import (
	"database/sql"
	"time"
)

type Form struct {
	Id              uint64
	Title           string
	Description     sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
	BackgroundColor sql.NullString
	BackgroundImage sql.NullString
	Image           sql.NullString
	Owner           uint64
	MultiResponse   bool
	Resubmission    bool
	Status          string
}

type FormQuestionOption struct {
	Caption string
	Value   sql.NullString
	Image   sql.NullString
}

type FormQuestion struct {
	Id            uint64
	Prompt        string
	Form          uint64
	IsRequired    bool
	Type          string
	LayoutVariant sql.NullString
	Options       []FormQuestionOption
}

type FormResponse struct {
	Id          uint64
	Responder   uint64
	SubmittedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FormResponseAnswer struct {
	Id        uint64
	Question  uint64
	Value     sql.NullString
	Response  uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}
