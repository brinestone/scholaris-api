package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Form statuses
const (
	FSDraft     = "draft"
	FSPublished = "published"
)

// Form question types
const (
	QTSingleline   string = "text"
	QTSingleChoice string = "single-choice"
	QTMCQ          string = "multiple-choice"
	QTFile         string = "file"
	QTDate         string = "date"
	QTGeoPoint     string = "coords"
	QTEmail        string = "email"
	QTMultiline    string = "multiline"
	QTTel          string = "tel"
)

var questionTypes = []string{QTSingleline, QTSingleChoice, QTMCQ, QTFile, QTDate, QTGeoPoint, QTEmail, QTMultiline, QTTel}

type FormAnswerUpdate struct {
	Question uint64  `json:"question"`
	Value    *string `json:"value,omitempty" encore:"optional"`
}

func (f FormAnswerUpdate) Validate() error {
	if f.Question == 0 {
		return fmt.Errorf("invalid question ID: %d", f.Question)
	}
	return nil
}

type UpdateUserAnswersRequest struct {
	Removed []uint64           `json:"removed,omitempty" encore:"optional"`
	Updated []FormAnswerUpdate `json:"updated,omitempty" encore:"optional"`
}

func (f UpdateUserAnswersRequest) Validate() error {
	msgs := make([]string, 0)

	if len(f.Removed) > 0 {
		for i, v := range f.Removed {
			if v == 0 {
				msgs = append(msgs, fmt.Sprintf("Invalid Answer ID at removed[%d]: %d", i, v))
			}
		}
	}

	if len(f.Updated) > 0 {
		for i, v := range f.Updated {
			if err := v.Validate(); err != nil {
				msgs = append(msgs, fmt.Sprintf("invalid update at updated[%d] - %s", i, err.Error()))
			}
		}
	}

	if len(f.Updated) == 0 && len(f.Removed) == 0 {
		msgs = append(msgs, "At least one of the fields updated,removed must be populated")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type FormAnswer struct {
	Id        uint64    `json:"id"`
	Question  uint64    `json:"question"`
	Value     *string   `json:"value,omitempty" encore:"optional"`
	CreatedAt time.Time `json:"answeredAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Response  uint64    `json:"response"`
}

type UserFormResponse struct {
	Id          uint64       `json:"id"`
	Responsder  uint64       `json:"responder"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	SubmittedAt *time.Time   `json:"submittedAt,omitempty" encore:"optional"`
	Answers     []FormAnswer `json:"answers,omitempty" encore:"optional"`
}

type UserFormResponses struct {
	Responses []UserFormResponse `json:"responses"`
}

type DeleteFormQuestionGroupsRequest struct {
	Ids []uint64 `json:"ids,omitempty"`
}

type UpdateFormQuestionGroupRequest struct {
	Label       *string `json:"label,omitempty" encore:"optional"`
	Description *string `json:"description,omitempty" encore:"optional"`
	Image       *string `json:"image,omitempty" encore:"optional"`
}

type DeleteQuestionsRequest struct {
	Questions []uint64 `json:"questions"`
}

func (d DeleteQuestionsRequest) Validate() error {
	msgs := make([]string, 0)
	if len(d.Questions) == 0 {
		msgs = append(msgs, "The questions field cannot be empty")
	} else {
		for i, id := range d.Questions {
			if id == 0 {
				msgs = append(msgs, fmt.Sprintf("Invalid question ID at questions[%d]", i))
			}
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type NewQuestionOption struct {
	Caption   string  `json:"caption"`
	Value     *string `json:"value,omitempty" encore:"optional"`
	IsDefault bool    `json:"isDefault"`
	Image     *string `json:"image,omitempty" encore:"optional"`
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
	Value     *string `json:"value,omitempty" encore:"optional"`
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
		return errors.New("there should be at least one entry in either the: added,removed or updates fields")
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
			} else {
				for j, w := range u.Removed {
					if w == v.Id {
						msgs = append(msgs, fmt.Sprintf("option marked for update at updates[%d] cannot be also marked for removal in removed[%d]", i, j))
					}
				}
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

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateFormQuestionRequest struct {
	Prompt        string  `json:"prompt"`
	IsRequired    bool    `json:"isRequired"`
	Type          string  `json:"type"`
	LayoutVariant *string `json:"layoutVariant"`
	Group         uint64  `json:"group,omitempty" encore:"optional"`
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
	Title           string         `json:"title,omitempty" encore:"optional"`
	Description     *string        `json:"description,omitempty" encore:"optional"`
	BackgroundColor *string        `json:"backgroundColor,omitempty" encore:"optional"`
	BackgroundImage *string        `json:"backgroundImage,omitempty" encore:"optional"`
	Image           *string        `json:"image,omitempty" encore:"optional"`
	MultiResponse   bool           `json:"multiResponse"`
	Resubmission    bool           `json:"resubmission"`
	CaptchaToken    string         `header:"x-ver-token"`
	Deadline        *time.Time     `json:"deadline,omitempty" encore:"optional"`
	MaxResponses    *uint          `json:"maxResponses,omitempty" encore:"optional"`
	MaxSubmissions  *uint          `json:"maxSubmissions,omitempty" encore:"optional"`
	ResponseStart   time.Time      `json:"responseStart"`
	ResponseWindow  *time.Duration `json:"responseWindow,omitempty" encore:"optional"`
}

func (n UpdateFormRequest) Validate() error {
	var msgs = make([]string, 0)

	if n.ResponseWindow != nil && n.ResponseWindow.Hours() <= 0 {
		msgs = append(msgs, "The value of the repsonseWindow if defined, must be greater than zero")
	}

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
	Id        uint64  `json:"id"`
	Caption   string  `json:"caption"`
	Value     *string `json:"value,omitempty" encore:"optional"`
	Image     *string `json:"image,omitempty" encore:"optional"`
	IsDefault bool    `json:"isDefault"`
}

type FormQuestion struct {
	Id            uint64           `json:"id"`
	Prompt        string           `json:"prompt"`
	Type          string           `json:"type"`
	IsRequired    bool             `json:"isRequired"`
	LayoutVariant string           `json:"layoutVariant,omitempty" encore:"optional"`
	Options       []QuestionOption `json:"options,omitempty" encore:"optional"`
	Group         uint64           `json:"group,omitempty" encore:"optional"`
}

type FormQuestionGroup struct {
	Id          uint64  `json:"id"`
	Label       *string `json:"label,omitempty" encore:"optional"`
	Form        uint64  `json:"form"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
}

type GetFormsResponse struct {
	Forms []FormConfig `json:"forms"`
}

type GetFormQuestionsResponse struct {
	Questions []FormQuestion      `json:"questions"`
	Groups    []FormQuestionGroup `json:"groups"`
}

type FindFormsRequest struct {
	Page      int    `query:"page"`
	Size      int    `query:"size"`
	Owner     uint64 `header:"x-owner"`
	OwnerType string `header:"x-owner-type"`
}

func (i FindFormsRequest) Validate() error {
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

type NewFormResponse struct {
	Id uint64 `json:"id"`
}

type NewFormInput struct {
	Title           string         `json:"title"`
	Description     *string        `json:"description,omitempty" encore:"optional"`
	BackgroundColor *string        `json:"backgroundColor,omitempty" encore:"optional"`
	BackgroundImage *string        `json:"backgroundImage,omitempty" encore:"optional"`
	Image           *string        `json:"image,omitempty" encore:"optional"`
	MultiResponse   bool           `json:"multiResponse"`
	Resubmission    bool           `json:"resubmission"`
	CaptchaToken    string         `header:"x-ver-token"`
	Owner           uint64         `header:"x-owner"`
	OwnerType       string         `header:"x-owner-type"`
	ResponseWindow  *time.Duration `json:"repsonseWindow,omitempty" encore:"optional"`
	ResponseStart   time.Time      `json:"responseStart"`
	MaxResponses    uint           `json:"maxResponses"`
	MaxSubmissions  uint           `json:"maxSubmissions"`
	Tags            []string       `json:"tags,omitempty" encore:"optional"`
}

func (n NewFormInput) GetCaptchaToken() string {
	return n.CaptchaToken
}

func (n NewFormInput) Validate() error {
	var msgs = make([]string, 0)

	if n.ResponseStart.Before(time.Now()) {
		msgs = append(msgs, "The responseStart value must be a future date")
	}

	if n.ResponseWindow != nil && n.ResponseWindow.Hours() == 0 {
		msgs = append(msgs, "The value for responseWindow cannot be zero")
	}

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
		msgs = append(msgs, "The x-owner header is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

// A form's current configuration.
type FormConfig struct {
	// The form's ID
	Id uint64 `json:"id"`
	// The form's title
	Title string `json:"title"`
	// A short description of the form
	Description *string `json:"description,omitempty" encore:"optional"`
	// A background color of the form
	BackgroundColor *string `json:"backgroundColor,omitempty" encore:"optional"`
	// The status of the form. Possible values are: **draft**, **published**. When published, the form is visible to everyone.
	Status string `json:"status"`
	// A background image URL of the form
	BackgroundImage *string `json:"backgroundImage,omitempty" encore:"optional"`
	// An image URL for the form.
	Image *string `json:"image,omitempty" encore:"optional"`
	// Whether a user can make multiple responses of the form.
	MultiResponse bool `json:"multiResponse"`
	// Whether a user can re-submit their a response to the form.
	Resubmission bool `json:"resubmission"`
	// The form's creation date.
	CreatedAt time.Time `json:"createdAt"`
	// The form's last modified date.
	UpdateAt time.Time `json:"updatedAt"`
	// An optional deadline for all response submissions of the form.
	Deadline *time.Time `json:"deadline,omitempty" encore:"optional"`
	// The maximum number of responses a user can make for the form.
	MaxResponses *uint `json:"maxResponses,omitempty" encore:"optional"`
	// The maxiumum number of submissions a user can make for the form.
	MaxSubmissions *uint `json:"maxSubmissions,omitempty" encore:"optional"`
	// Tags attached to the form
	Tags           []string       `json:"tags,omitempty"`
	GroupIds       []uint64       `json:"groupIds,omitempty"`
	QuestionIds    []uint64       `json:"questionIds,omitempty"`
	ResponseStart  time.Time      `json:"responseStart"`
	ResponseWindow *time.Duration `json:"responseWindow,omitempty" encore:"optional"`
}
