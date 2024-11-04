package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	QTSingleline   string = "text"
	QTSingleChoice string = "single-choice"
	QTMCQ          string = "multiple-choice"
	QTFile         string = "file"
	QTDate         string = "date"
	QTGeoPoint     string = "coords"
	QTEmail        string = "email"
	QTMultiline    string = "multi-line"
	QTTel          string = "tel"
)

var questionTypes = []string{QTSingleline, QTSingleChoice, QTMCQ, QTFile, QTDate, QTGeoPoint, QTEmail, QTMultiline, QTTel}

type NewQuestionOption struct {
	Caption   string  `json:"caption"`
	Value     *string `json:"value,omitempty"`
	IsDefault bool    `json:"isDefault"`
	Image     *string `json:"image,omitempty"`
}

func (n NewQuestionOption) Validate() error {
	if len(n.Caption) == 0 {
		return errors.New("the caption field is required")
	}
	return nil
}

type FormQuestionOptionUpdate struct {
	Id        uint64  `json:"id"`
	Caption   string  `json:"caption"`
	Value     *string `json:"value,omitempty"`
	Image     *string `json:"image"`
	IsDefault bool    `json:"isDefault"`
}

func (f FormQuestionOptionUpdate) Validate() error {
	msgs := make([]string, 0)

	if f.Id == 0 {
		msgs = append(msgs, "Invalid value for id")
	}

	if len(f.Caption) == 0 {
		msgs = append(msgs, "The caption field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateFormQuestionOptionsRequest struct {
	Removed []uint64                   `json:"removed"`
	Added   []NewQuestionOption        `json:"added"`
	Updates []FormQuestionOptionUpdate `json:"updates"`
}

func (u UpdateFormQuestionOptionsRequest) Validate() error {
	if len(u.Added) == 0 && len(u.Removed) == 0 && len(u.Updates) == 0 {
		return errors.New("there should be at least one entry in all the fields")
	}

	msgs := make([]string, 0)

	if len(u.Added) > 0 {
		for i, v := range u.Added {
			if err := v.Validate(); err != nil {
				msgs = append(msgs, fmt.Sprintf("error at added[%d] - %s", i, err.Error()))
			}
		}
	}

	if len(u.Updates) > 0 {
		for i, v := range u.Updates {
			if err := v.Validate(); err != nil {
				msgs = append(msgs, fmt.Sprintf("error at updates[%d] - %s", i, err.Error()))
			}
		}
	}

	if len(u.Removed) > 0 {
		for i, v := range u.Removed {
			if v == 0 {
				msgs = append(msgs, fmt.Sprintf("Invalid ID at removed[%d]", i))
			}
		}
	}

	if len(msgs) == 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateFormQuestionRequest struct {
	Prompt        string  `json:"prompt"`
	IsRequired    bool    `json:"isRequired"`
	Type          string  `json:"type"`
	LayoutVariant *string `json:"layoutVariant"`
	// Options       []NewQuestionOption `json:"options"`
}

func (n UpdateFormQuestionRequest) Validate() error {
	msgs := make([]string, 0)

	if len(n.Type) > 0 {
		isValid := false
		for _, k := range questionTypes {
			if k == n.Type {
				isValid = true
				break
			}
		}
		if !isValid {
		}
	} else {
		msgs = append(msgs, "Invalid question type")
	}

	if len(n.Prompt) <= 0 {
		msgs = append(msgs, "The prompt field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateFormRequest struct {
	// AddedQuestions   []FormQuestion `json:"addedQuestions,omitempty"`
	// RemovedQuestions []uint64       `json:"removedQuestions,omitempty"`
	Title           string  `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	BackgroundColor *string `json:"backgroundColor,omitempty"`
	BackgroundImage *string `json:"backgroundImage,omitempty"`
	Image           *string `json:"image,omitempty"`
	MultiResponse   bool    `json:"multiResponse"`
	Resubmission    bool    `json:"resubmission"`
	CaptchaToken    string  `header:"x-ver-token"`
}

func (n UpdateFormRequest) Validate() error {
	var msgs = make([]string, 0)

	if len(n.Title) == 0 {
		msgs = append(msgs, "The title field is required")
	}

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token failed")
	}

	if n.BackgroundColor != nil && len(*n.BackgroundColor) > 0 && len(*n.BackgroundColor) < 7 {
		msgs = append(msgs, "Invalid background color")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}

	return nil
}

type QuestionOption struct {
	Id      uint64  `json:"id"`
	Caption string  `json:"caption"`
	Value   *string `json:"value,omitempty"`
	Image   *string `json:"image,omitempty"`
}

type FormQuestion struct {
	Id            uint64           `json:"id"`
	Prompt        string           `json:"prompt"`
	Type          string           `json:"type"`
	IsRequired    bool             `json:"isRequired"`
	LayoutVariant string           `json:"layoutVariant"`
	Options       []QuestionOption `json:"options,omitempty"`
}

type GetFormQuestionsResponse struct {
	Questions []FormQuestion `json:"questions"`
}

type GetFormsInput struct {
	Page      int    `query:"page"`
	Size      int    `query:"size"`
	Owner     uint64 `header:"x-owner"`
	OwnerType string `header:"x-owner-type"`
}

func (i GetFormsInput) Validate() error {
	var msgs []string
	if len(i.OwnerType) == 0 {
		msgs = append(msgs, "x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(i.OwnerType); !ok {
		msgs = append(msgs, "Invalid owner type")
	}

	if i.Owner == 0 {
		msgs = append(msgs, "The x-owner field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type NewFormInput struct {
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	BackgroundColor *string `json:"backgroundColor,omitempty"`
	BackgroundImage *string `json:"backgroundImage,omitempty"`
	Image           *string `json:"image,omitempty"`
	MultiResponse   bool    `json:"multiResponse"`
	Resubmission    bool    `json:"resubmission"`
	CaptchaToken    string  `header:"x-ver-token"`
	Owner           uint64  `header:"x-owner"`
	OwnerType       string  `header:"x-owner-type"`
}

func (n NewFormInput) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewFormInput) Validate() error {
	var msgs = make([]string, 0)

	if len(n.Title) == 0 {
		msgs = append(msgs, "The title field is required")
	}

	if len(n.CaptchaToken) == 0 {
		msgs = append(msgs, "Invalid captcha token failed")
	}

	if len(n.OwnerType) == 0 {
		msgs = append(msgs, "x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(n.OwnerType); !ok {
		msgs = append(msgs, "Invalid owner type")
	}

	if n.Owner == 0 {
		msgs = append(msgs, "The x-owner field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type FormConfig struct {
	Id              uint64    `json:"id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	BackgroundColor *string   `json:"backgroundColor,omitempty"`
	Status          string    `json:"status"`
	BackgroundImage *string   `json:"backgroundImage,omitempty"`
	Image           *string   `json:"image,omitempty"`
	MultiResponse   bool      `json:"multiResponse"`
	Resubmission    bool      `json:"resubmission"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdateAt        time.Time `json:"updatedAt"`
}
