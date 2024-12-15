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
	case string(PNCanViewInstitutions):
		return PNCanViewInstitutions, true
	case string(PNOwner):
		return PNOwner, true
	case string(PNParent):
		return PNParent, true
	case string(PNMember):
		return PNMember, true
	case string(PNDestination):
		return PNDestination, true
	case string(PNCanCreateForms):
		return PNCanCreateForms, true
	case string(PNFormResponder):
		return PNFormResponder, true
	case string(PNCanView):
		return PNCanView, true
	case string(PNCanViewSettings):
		return PNCanViewSettings, true
	case string(PNCanEditSettings):
		return PNCanEditSettings, true
	case string(PNEditor):
		return PNEditor, true
	case string(PNCanEdit):
		return PNCanEdit, true
	case string(PNCanCreateAcademicYear):
		return PNCanCreateAcademicYear, true
	case string(PNCanEnroll):
		return PNCanEnroll, true
	case string(PNCanCreateInstitution):
		return PNCanCreateInstitution, true
	case string(PNCanModifyMembers):
		return PNCanModifyMembers, true
	case string(PNCanViewMembers):
		return PNCanViewMembers, true
	case string(PNCanSetSettingValue):
		return PNCanSetSettingValue, true
	case string(PNCanUploadFile):
		return PNCanUploadFile, true
	case string(PNCanChangeOwner):
		return PNCanChangeOwner, true
	case string(PNCanChangeSettings):
		return PNCanChangeSettings, true
	case string(PNCanDelete):
		return PNCanDelete, true
	case string(PNCanUpdate):
		return PNCanUpdate, true
	case string(PNCanUpdateSubscription):
		return PNCanUpdateSubscription, true
	case string(PNCanCreateSettings):
		return PNCanCreateSettings, true
	default:
		return pnUnknown, false
	}
}

// Permission name
const (
	PNOwner                 PermissionName = "owner"
	PNParent                PermissionName = "parent"
	PNMember                PermissionName = "member"
	PNDestination           PermissionName = "destination"
	PNCanCreateAcademicYear PermissionName = "can_create_academic_year"
	PNCanCreateInstitution  PermissionName = "can_create_institution"
	PNCanCreateForms        PermissionName = "can_create_forms"
	PNFormResponder         PermissionName = "responder"
	PNCanView               PermissionName = "can_view"
	PNCanViewSettings       PermissionName = "can_view_settings"
	PNCanUploadFile         PermissionName = "can_upload_file"
	PNCanSetSettingValue    PermissionName = "can_set_setting_value"
	PNCanEditSettings       PermissionName = "can_edit_settings"
	PNEditor                PermissionName = "editor"
	PNCanEdit               PermissionName = "can_edit"
	PNCanEnroll             PermissionName = "can_enroll"
	PNCanModifyMembers      PermissionName = "can_modify_members"
	PNCanViewMembers        PermissionName = "can_view_members"
	PNCanChangeOwner        PermissionName = "can_change_owner"
	PNCanChangeSettings     PermissionName = "can_change_settings"
	PNCanDelete             PermissionName = "can_delete"
	PNCanUpdate             PermissionName = "can_update"
	PNCanUpdateSubscription PermissionName = "can_update_subscription"
	PNCanCreateSettings     PermissionName = "can_create_settings"
	PNCanViewInstitutions   PermissionName = "can_view_institutions"
	pnUnknown               PermissionName = ""
)

type ListObjectsResponse struct {
	// The valid relations
	Relations map[PermissionType][]uint64 `json:"relations"`
}

type ListRelationsResponse struct {
	Relations []string `json:"relations"`
}

type ListRelationsRequest struct {
	Permissions []string `json:"permissions"`
	Target      string   `json:"target"`
	// Context   *map[string]any `json:"context,omitempty" encore:"optional"`
}

func (l ListRelationsRequest) Validate() (err error) {
	msgs := make([]string, 0)

	if len(l.Permissions) == 0 {
		msgs = append(msgs, "The permissions field is required and cannot be empty")
	} else {
		for i, v := range l.Permissions {
			_, validName := ParsePermissionName(v)
			if !validName {
				msgs = append(msgs, fmt.Sprintf("Invalid permission name at permissions[%d]", i))
			}
		}
	}

	if len(msgs) > 0 {
		err = errors.New(strings.Join(msgs, "\n"))
	}
	return
}

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
