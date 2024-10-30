package models

import (
	"database/sql"
	"time"
)

type EnrollmentFormQuestionOption struct {
	IsDefault bool   `json:"isDefault"`
	Value     string `json:"value"`
	Label     string `json:"label"`
}

type EnrollmentFormQuestion struct {
	Id              uint64
	Institution     uint64
	QuestionType    sql.NullString
	AnswerType      sql.NullString
	IsRequired      sql.NullBool
	ChoiceDelimiter sql.NullString
	Prompt          string
	Options         []EnrollmentFormQuestionOption
}

type EnrollmentFormAnswer struct {
	Value      []sql.NullString `json:"value"`
	AnsweredAt time.Time        `json:"answeredAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
	QuestionId uint64           `json:"question"`
}

type Enrollment struct {
	Id                 uint64
	Owner              uint64
	Destination        uint64
	Status             string
	Approver           sql.NullInt64
	ApprovedAt         sql.NullTime
	PaymentTransaction sql.NullInt64
	ServiceTransaction sql.NullInt64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Documents          []string
	Answers            []EnrollmentFormAnswer
}
