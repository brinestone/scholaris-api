package dto

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"encore.dev/beta/auth"
)

type PermissionType string
type PermissionName string

func IdentifierString[T auth.UID | uint64 | string](pt PermissionType, id T) string {
	return fmt.Sprintf("%s:%v", pt, id)
}

func ParsePermissionType(s string) (PermissionType, bool) {
	switch s {
	case string(PTInstitution):
		return PTInstitution, true
	case string(PTUser):
		return PTUser, true
	case string(PTEnrollment):
		return PTEnrollment, true
	case string(PTTenant):
		return PTTenant, true
	case string(PTForm):
		return PTForm, true
	case string(PTSubscription):
		return PTSubscription, true
	case string(PTSetting):
		return PTSetting, true
	case string(PTAcademicYear):
		return PTAcademicYear, true
	case string(PTAcademicTerm):
		return PTAcademicTerm, true
	case string(PTUserFile):
		return PTUserFile, true
	case string(PTSharedFile):
		return PTSharedFile, true
	default:
		return unknown, false
	}
}

const (
	PTInstitution  PermissionType = "institution"
	PTUser         PermissionType = "user"
	PTTenant       PermissionType = "tenant"
	PTForm         PermissionType = "form"
	PTEnrollment   PermissionType = "enrollment"
	PTSubscription PermissionType = "subscription"
	PTSetting      PermissionType = "setting"
	PTAcademicYear PermissionType = "academicYear"
	PTAcademicTerm PermissionType = "academicTerm"
	PTUserFile     PermissionType = "file"
	PTSharedFile   PermissionType = "shared_file"
	unknown        PermissionType = ""
)

func ParsePermissionName(p string) (PermissionName, bool) {
	switch p {
	case string(PermOwner):
		return PermOwner, true
	case string(PermParent):
		return PermParent, true
	case string(PermMember):
		return PermMember, true
	case string(PermDestination):
		return PermDestination, true
	case string(PermCanCreateForms):
		return PermCanCreateForms, true
	case string(PermFormResponder):
		return PermFormResponder, true
	case string(PermCanView):
		return PermCanView, true
	case string(PermCanViewSettings):
		return PermCanViewSettings, true
	case string(PermCanEditSettings):
		return PermCanEditSettings, true
	case string(PermEditor):
		return PermEditor, true
	case string(PermCanEdit):
		return PermCanEdit, true
	case string(PermCanCreateAcademicYear):
		return PermCanCreateAcademicYear, true
	case string(PermCanEnroll):
		return PermCanEnroll, true
	case string(PermCanCreateInstitution):
		return PermCanCreateInstitution, true
	default:
		return permUnknown, false
	}
}

// Permission name
const (
	PermOwner                 PermissionName = "owner"
	PermParent                PermissionName = "parent"
	PermMember                PermissionName = "member"
	PermDestination           PermissionName = "destination"
	PermCanCreateAcademicYear PermissionName = "can_create_academic_year"
	PermCanCreateInstitution  PermissionName = "can_create_institution"
	PermCanCreateForms        PermissionName = "can_create_forms"
	PermFormResponder         PermissionName = "responder"
	PermCanView               PermissionName = "can_view"
	PermCanViewSettings       PermissionName = "can_view_settings"
	PermCanUploadFile         PermissionName = "can_upload_file"
	PermCanSetSettingValue    PermissionName = "can_set_setting_value"
	PermCanEditSettings       PermissionName = "can_edit_settings"
	PermEditor                PermissionName = "editor"
	PermCanEdit               PermissionName = "can_edit"
	PermCanEnroll             PermissionName = "can_enroll"
	permUnknown               PermissionName = ""
)

type ListObjectsResponse struct {
	// The valid relations
	Relations map[PermissionType][]uint64 `json:"relations"`
}

// type ListRelationsResponse struct {
// 	Relations []string `json:"relations"`
// }

// type ListRelationsRequest struct {
// 	Roles  []string `json:"roles"`
// 	Target string   `json:"target"`
// 	// Context   *map[string]any `json:"context,omitempty" encore:"optional"`
// }

// func (l ListRelationsRequest) Validate() (err error) {
// 	msgs := make([]string, 0)

// 	if len(l.Roles) == 0 {
// 		msgs = append(msgs, "The relations field is required")
// 	} else {
// 		pattern := regexp.MustCompile(`^[a-zA-Z_\-0-9]+:[a-zA-Z_\-0-9]+$`)
// 		fn := func(rel string) bool {
// 			patternMatches := pattern.MatchString(rel)
// 			_, validType := ParsePermissionType(strings.Split(rel, ":")[0])
// 			return patternMatches && validType
// 		}

// 		valid := helpers.Every(l.Roles, fn)
// 		if !valid {
// 			msgs = append(msgs, "Erroneous relation specifier detected")
// 		}
// 	}

// 	if len(msgs) > 0 {
// 		err = errors.New(strings.Join(msgs, "\n"))
// 	}
// 	return
// }

type ListObjectsRequest struct {
	// The object claiming to own the relation.
	Actor string `json:"subject"`
	// The relation specifier
	Relation PermissionName `json:"relation"`
	// The target object
	Type    string         `json:"type"`
	Context []ContextEntry `json:"context,omitempty" encore:"optional"`
}

type BatchRelationCheckResponse struct {
	Results map[string]bool `json:"results"`
}

type RelationCheckResponse struct {
	Allowed bool `json:"allowed"`
}

type InternalRelationCheckRequest struct {
	// The actor's identifier who owns the relation
	Actor string `json:"actor"`
	// The relation specicfier
	Relation PermissionName `json:"relation"`
	// The target resource identifier
	Target    string             `json:"target"`
	Condition *RelationCondition `json:"condition,omitempty" encore:"optional"`
}

type RelationCheck struct {
	// The relation specicfier
	Relation string `json:"relation"`
	// The target resource identifier
	Target string `json:"target"`
}

func (r RelationCheck) Validate() (err error) {
	msgs := make([]string, 0)

	_, valid := ParsePermissionName(r.Relation)
	if !valid {
		msgs = append(msgs, "Invalid permission name")
	}

	if len(r.Target) == 0 {
		msgs = append(msgs, "Invalid target value")
	} else {
		pattern := regexp.MustCompile(`^[a-zA-Z_\-0-9]+:[a-zA-Z_\-0-9]+$`)
		patternMatches := pattern.MatchString(r.Target)
		_, validType := ParsePermissionType(strings.Split(r.Target, ":")[0])
		if !(patternMatches && validType) {
			msgs = append(msgs, "Invalid value for \"target\"")
		}
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}
	return
}

type BatchRelationCheckRequest struct {
	Checks []RelationCheck `json:"checks"`
}

func (b BatchRelationCheckRequest) Validate() error {
	msgs := make([]string, 0)

	for i, v := range b.Checks {
		err := v.Validate()
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("invalid check at checks[%d] - %s", i, err.Error()))
		}
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type ContextEntry struct {
	Name  string
	Type  ContextEntryType
	Value string
}

type RelationCondition struct {
	Name    string
	Context []ContextEntry
}

type ContextEntryType string

const (
	CETTimestamp ContextEntryType = "timestamp"
	CETBool      ContextEntryType = "bool"
	CETString    ContextEntryType = "string"
	CETDuration  ContextEntryType = "duration"
)

func HavingEntry(name string, _type ContextEntryType, value any) ContextEntry {
	return ContextEntry{
		Name:  name,
		Type:  _type,
		Value: fmt.Sprintf("%v", value),
	}
}

func WithCondition(name string, entries ...ContextEntry) RelationCondition {
	ans := RelationCondition{
		Name:    name,
		Context: make([]ContextEntry, len(entries)),
	}
	copy(ans.Context, entries)

	return ans
}

type PermissionUpdate struct {
	// The actor who owns the relation
	Actor string
	// The relation specifier
	Relation PermissionName
	// The target resource identifier
	Target string
	// The conditions of the relation
	Condition *RelationCondition
}

func (p PermissionUpdate) WithCondition(c *RelationCondition) PermissionUpdate {
	p.Condition = c
	return p
}

func NewPermissionUpdate[T string | uint64 | auth.UID](actor string, relation PermissionName, target string) PermissionUpdate {
	return NewPermissionUpdateWithCondition[T](actor, relation, target, nil)
}

func NewPermissionUpdateWithCondition[T string | uint64 | auth.UID](actor string, relation PermissionName, target string, cond *RelationCondition) PermissionUpdate {
	return PermissionUpdate{
		Actor:     actor,
		Relation:  relation,
		Target:    target,
		Condition: cond,
	}
}

type UpdatePermissionsRequest struct {
	Updates []PermissionUpdate
}
