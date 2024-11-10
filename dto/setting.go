package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type SetValue struct {
	Value *string `json:"value,omitempty" encore:"optional"`
	Index *uint   `json:"id,omitempty" encore:"optional"`
}

type SettingValueUpdate struct {
	Key   string     `json:"key"`
	Value []SetValue `json:"value,omitempty" encore:"optional"`
}

func (s SettingValueUpdate) Validate() error {
	if len(s.Key) == 0 {
		return errors.New("the key field is required")
	}
	return nil
}

type SetSettingValueRequest struct {
	Owner     uint64               `header:"x-owner"`
	OwnerType string               `header:"x-owner-type"`
	Updates   []SettingValueUpdate `json:"updates"`
}

func (s SetSettingValueRequest) Validate() error {
	msgs := make([]string, 0)
	if s.Owner == 0 {
		msgs = append(msgs, "The x-owner header is required")
	}

	if len(s.OwnerType) == 0 {
		msgs = append(msgs, "The x-owner-type header is required")
	} else if _, ok := PermissionTypeFromString(s.OwnerType); !ok {
		msgs = append(msgs, "The x-owner-type header value is invalid")
	}

	if len(s.Updates) == 0 {
		msgs = append(msgs, "The updates field cannot be empty")
	} else {
		for i, v := range s.Updates {
			if err := v.Validate(); err != nil {
				msgs = append(msgs, fmt.Sprintf("Error at updates[%d] - %s", i, err.Error()))
			}
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

func (s SetSettingValueRequest) GetOwner() uint64 {
	return s.Owner
}

func (s SetSettingValueRequest) GetOwnerType() string {
	return s.OwnerType
}

type SettingOption struct {
	Id      uint64  `json:"id"`
	Label   string  `json:"label"`
	Value   *string `json:"value"`
	Setting uint64  `json:"setting"`
}

type SettingValue struct {
	Id      uint64     `json:"id"`
	Setting uint64     `json:"setting"`
	SetAt   *time.Time `json:"setAt,omitempty" encore:"optional"`
	SetBy   uint64     `json:"setBy"`
	Value   *string    `json:"value,omitempty" encore:"optional"`
	Index   uint       `json:"index"`
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
	Key         string  `json:"key,omitempty"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty" encore:"optional"`
	MultiValues bool    `json:"multiValues"`
	// SystemGenerated bool                  `json:"systemGenerated"`
	Parent      *uint64               `json:"parent,omitempty" encore:"optional"`
	Options     []SettingOptionUpdate `json:"options,omitempty" encore:"optional"`
	Overridable bool                  `json:"overrridable"`
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

	if len(s.Key) == 0 {
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
	Deletes      []string        `json:"deletes,omitempty" encore:"optional"`
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

	if len(u.Deletes) > 0 {
		for i, v := range u.Deletes {
			if len(v) == 0 {
				msgs = append(msgs, fmt.Sprintf("Invalid value at deletes[%d]", i))
			}
		}
	}

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

	if len(u.Updates) > 0 && len(u.Deletes) > 0 {
		for i := 0; i < len(u.Updates); i++ {
			for j := 0; j < len(u.Deletes); j++ {
				if u.Deletes[j] == u.Updates[i].Key {
					msgs = append(msgs, fmt.Sprintf("Key \"%s\" cannot be marked for update (at updates[%d]) and for deletion (at deletes[%d])", u.Updates[i].Key, i, j))
				}
			}
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type GetSettingsRequest struct {
	Owner     uint64 `header:"x-owner"`
	OwnerType string `header:"x-owner-type"`
}

func (g GetSettingsRequest) GetOwnerType() string {
	return g.OwnerType
}

func (g GetSettingsRequest) GetOwner() uint64 {
	return g.Owner
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
