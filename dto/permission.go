package dto

import (
	"fmt"

	"encore.dev/beta/auth"
)

type PermissionType string

func IdentifierString[T auth.UID | uint64 | string](pt PermissionType, id T) string {
	return fmt.Sprintf("%s:%v", pt, id)
}

func PermissionTypeFromString(s string) (PermissionType, bool) {
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
	unknown        PermissionType = ""
)

type ListRelationsResponse struct {
	// The valid relations
	Relations map[PermissionType][]uint64 `json:"relations"`
}

type ListRelationsRequest struct {
	// The object claiming to own the relation.
	Actor string `json:"subject"`
	// The relation specifier
	Relation string `json:"relation"`
	// The target object
	Type string `json:"type"`
}

type RelationCheckResponse struct {
	Allowed bool `json:"allowed"`
}

type RelationCheckRequest struct {
	// The actor's identifier who owns the relation
	Actor string `json:"actor"`
	// The relation specicfier
	Relation string `json:"relation"`
	// The target resource identifier
	Target    string             `json:"target"`
	Condition *RelationCondition `json:"condition,omitempty" encore:"optional"`
}

type ContextEntry struct {
	Name  string
	Type  string
	Value string
}

type RelationCondition struct {
	Name    string
	Context []ContextEntry
}

func HavingEntry(name, _type, value string) ContextEntry {
	return ContextEntry{
		Name:  name,
		Type:  _type,
		Value: value,
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
	Relation string
	// The target resource identifier
	Target string
	// The conditions of the relation
	Condition *RelationCondition
}

func (p PermissionUpdate) WithCondition(c *RelationCondition) PermissionUpdate {
	p.Condition = c
	return p
}

func NewPermissionUpdate[T string | uint64 | auth.UID](actor, relation, target string) PermissionUpdate {
	return NewPermissionUpdateWithCondition[T](actor, relation, target, nil)
}

func NewPermissionUpdateWithCondition[T string | uint64 | auth.UID](actor, relation, target string, cond *RelationCondition) PermissionUpdate {
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
