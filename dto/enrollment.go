package dto

import (
	"errors"
	"fmt"
	"strings"
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
	Options         []*EnrollmentQuestionOption `json:"options,omitempty" encore:"optional"`
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
		Value    string `json:"value,omitempty" encore:"optional"`
		Question uint64 `json:"question"`
	} `json:"answers,omitempty" encore:"optional"`
	RemovedAnswers   []int
	AddedDocuments   []string `json:"addedDocuments,omitempty" encore:"optional"`
	RemovedDocuments []string `json:"removedDocuments,omitempty" encore:"optional"`
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
		Value    []*string `json:"value,omitempty" encore:"optional"`
		Question uint64    `json:"question"`
	} `json:"answers,omitempty" encore:"optional"`
	Documents []string `json:"documents,omitempty" encore:"optional"`
}
