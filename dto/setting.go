package dto

import (
	"errors"
	"fmt"
	"strings"
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
	Setting uint64     `json:"setting"`
	SetAt   *time.Time `json:"setAt"`
	SetBy   uint64     `json:"setBy"`
}

type Setting struct {
	Id              uint64          `json:"id"`
	Label           string          `json:"label"`
	Description     *string         `json:"description,omitempty" encore:"optional"`
	Key             string          `json:"key"`
	MultiValues     bool            `json:"multiValues"`
	SystemGenerated bool            `json:"systemGenerated"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	Owner           uint64          `json:"owner"`
	OwnerType       string          `json:"ownerType"`
	Overrideable    bool            `json:"overridable"`
	CreatedBy       uint64          `json:"createdBy"`
	Options         []SettingOption `json:"options,omitempty"`
	Values          []SettingValue  `json:"values,omitempty"`
	Parent          *uint64         `json:"parent,omitempty" encore:"optional"`
}

type SettingOptionUpdate struct {
	Label string  `json:"label"`
	Value *string `json:"value,omitempty" encore:"optional"`
	Key   *string `json:"key,omitempty" encore:"optional"` // !The key of the setting
	//// Key   *string `json:"key,omitempty" encore:"optional"`
}

func (s SettingOptionUpdate) Validate() error {
	msgs := make([]string, 0)

	if len(s.Label) == 0 {
		msgs = append(msgs, "The label field is required")
	}

	if s.Key != nil && len(*s.Key) == 0 {
		msgs = append(msgs, "The key field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type SettingUpdate struct {
	Key             *string               `json:"key,omitempty" encore:"optional"`
	Label           string                `json:"label"`
	Description     *string               `json:"description,omitempty" encore:"optional"`
	MultiValues     bool                  `json:"multiValues"`
	SystemGenerated bool                  `json:"systemGenerated"`
	Parent          *uint64               `json:"parent,omitempty" encore:"optional"`
	Options         []SettingOptionUpdate `json:"options,omitempty" encore:"optional"`
	Overridable     bool                  `json:"overrridable"`
}

func (s SettingUpdate) Validate() error {
	msgs := make([]string, 0)

	if len(s.Label) == 0 {
		msgs = append(msgs, "The label field is required")
	}

	if s.Parent != nil && *s.Parent == 0 {
		msgs = append(msgs, "The parent field is invalid")
	}

	if len(s.Options) > 0 {
		for i, v := range s.Options {
			if err := v.Validate(); err != nil {
				msgs = append(msgs, fmt.Sprintf("Error at options[%d] - %s", i, err.Error()))
			}
		}
	}

	if s.Key != nil && len(*s.Key) == 0 {
		msgs = append(msgs, "The key field is required")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type UpdateSettingsRequest struct {
	OwnerType    string          `header:"x-owner-type"`
	CaptchaToken string          `header:"x-ver-token"`
	Owner        uint64          `header:"x-owner"`
	Updates      []SettingUpdate `json:"updates"`
}

func (u UpdateSettingsRequest) GetOwnerType() string {
	return u.OwnerType
}

func (u UpdateSettingsRequest) GetOwner() uint64 {
	return u.Owner
}

func (u UpdateSettingsRequest) GetCaptchaToken() string {
	return u.CaptchaToken
}

func (u UpdateSettingsRequest) Validate() error {
	msgs := make([]string, 0)

	if u.Owner == 0 {
		msgs = append(msgs, "The x-owner header is required")
	}

	if len(u.CaptchaToken) == 0 {
		msgs = append(msgs, "The x-ver-token header is required")
	}

	if len(u.OwnerType) == 0 {
		msgs = append(msgs, "The x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(u.OwnerType); !ok {
		msgs = append(msgs, "The x-owner-type header value is invalid")
	}

	if len(u.Updates) == 0 {
		msgs = append(msgs, "The updates field cannot be empty")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type GetSettingsRequest struct {
	Owner     uint64 `header:"x-owner"`
	OwnerType string `header:"x-owher-type"`
}

func (g GetSettingsRequest) Validate() error {
	msgs := make([]string, 0)

	if g.Owner == 0 {
		msgs = append(msgs, "The x-owner header is required")
	}

	if len(g.OwnerType) == 0 {
		msgs = append(msgs, "The x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(g.OwnerType); !ok {
		msgs = append(msgs, "The x-owner-type header value is invalid")
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type GetSettingsResponse struct {
	Settings []Setting `json:"settings,omitempty"`
}
