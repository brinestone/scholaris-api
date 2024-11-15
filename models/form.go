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
	Deadline        sql.NullTime
	MaxResponses    sql.NullInt32
	MaxSubmissions  sql.NullInt32
	OwnerType       string
	QuestionIds     []uint64
	GroupIds        []uint64
	ResponseCount   uint64
	Tags            []string
	SubmissionCount uint64
}

type FormQuestionOption struct {
	Id        uint64
	Caption   string
	Value     *string
	IsDefault bool
	Image     *string
}

type FormQuestion struct {
	Id            uint64
	Prompt        string
	Form          uint64
	IsRequired    bool
	Type          string
	LayoutVariant sql.NullString
	Group         sql.NullInt64
	Options       []FormQuestionOption
}

type FormQuestionGroup struct {
	Id          uint64
	Label       sql.NullString
	Description sql.NullString
	Image       sql.NullString
	Form        uint64
}

type FormResponse struct {
	Id          uint64
	Responder   uint64
	SubmittedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Form        uint64
	Answers     []FormResponseAnswer
}

type FormResponseAnswer struct {
	Id        uint64    `json:"id"`
	Question  uint64    `json:"question"`
	Value     *string   `json:"value,omitempty" encore:"optional"`
	Response  uint64    `json:"response"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
