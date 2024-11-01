package dto

import (
	"errors"
	"fmt"
	"strings"
)

// Form Answer types
const (
	FAText           = "text"
	FASingleChoice   = "single-choice"
	FAMultipleChoice = "multiple-choice"
	FAFile           = "file"
)

// Question types
const (
	// Open ended question, meaning the user can provide any text as answer
	QTOpenEnded = "open-ended"
	// The user can provide multiple texts as answer
	QTMultipleChoice = "multiple-choice"
	// The user can only provide a set of URLs as answer
	QTUpload = "upload"
)

// Enrollment statuses
const (
	// The enrollment is in draft mode.
	ESDraft = "draft"
	// The enrollment has pending.
	ESPending = "pending"
	// The enrollment has been rejected.
	ESRejected = "rejected"
	// The enrollment has been approved.
	ESApproved = "approved"
)

type EnrollmentQuestionOption struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	IsDefault bool   `json:"isDefault"`
}

type EnrollmentQuestion struct {
	Id              uint64                      `json:"id"`
	Prompt          string                      `json:"prompt"`
	QuestionType    string                      `json:"questionType"`
	AnswerType      string                      `json:"answerType"`
	IsRequired      bool                        `json:"isRequired"`
	ChoiceDelimiter string                      `json:"choiceDelimiter"`
	Options         []*EnrollmentQuestionOption `json:"options,omitempty"`
}

type EnrollmentQuestions struct {
	Questions []*EnrollmentQuestion `json:"questions"`
}

type NewEnrollment struct {
	Destination             string `json:"institution"`
	ServiceTransactionToken string `json:"serviceTransactionToken"`
}

func (ne NewEnrollment) Validate() error {
	var msgs = make([]string, 0)
	if len(ne.Destination) == 0 {
		msgs = append(msgs, "Invalid institution ID")
	}

	if len(ne.ServiceTransactionToken) == 0 {
		msgs = append(msgs, "Invalid Service Transaction Verification Token")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateEnrollment struct {
	Destination uint64 `json:"institution"`
	Answers     []struct {
		Value    string `json:"value,omitempty"`
		Question uint64 `json:"question"`
	} `json:"answers,omitempty"`
	RemovedAnswers   []int
	AddedDocuments   []string `json:"addedDocuments,omitempty"`
	RemovedDocuments []string `json:"removedDocuments,omitempty"`
}

func (ue UpdateEnrollment) Validate() error {
	var msgs = make([]string, 0)
	if ue.Destination == 0 {
		msgs = append(msgs, "invalid institution ID")
	}

	if len(ue.Answers) > 0 {
		for i, a := range ue.Answers {
			if a.Question == 0 {
				msgs = append(msgs, fmt.Sprintf("Invalid Question at %d", i))
			}
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type EnrollmentState struct {
	Id          uint64 `json:"id"`
	Destination uint64 `json:"institution"`
	Answers     []struct {
		Value    []*string `json:"value,omitempty"`
		Question uint64    `json:"question"`
	} `json:"answers,omitempty"`
	Documents []string `json:"documents,omitempty"`
}
