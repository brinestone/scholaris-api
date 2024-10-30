package dto

import (
	"fmt"

	"encore.dev/beta/auth"
)

type PermissionType string

func IdentifierString[T auth.UID | uint64](pt PermissionType, id T) string {
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
	case string(PTSubscription):
		return PTSubscription, true
	default:
		return unknown, false
	}
}

const (
	PTInstitution  PermissionType = "institution"
	PTUser         PermissionType = "user"
	PTTenant       PermissionType = "tenant"
	PTEnrollment   PermissionType = "enrollment"
	PTSubscription PermissionType = "subscription"
	unknown        PermissionType = ""
)

type ListRelationsResponse struct {
	// The valid relations
	Relations map[PermissionType][]string `json:"relations"`
}

type ListRelationsRequest struct {
	// The object claiming to own the relation.
	Subject string `json:"subject"`
	// The relation specifier
	Relation string `json:"relation"`
	// The target object
	Type string `json:"type"`
}

type RelationCheckResponse struct {
	Allowed bool `json:"allowed"`
}

type RelationCheckRequest struct {
	Subject  string `query:"subject"`
	Relation string `query:"relation"`
	Target   string `query:"target"`
}

// func (r *RelationCheckRequest) From(val any) {

// }

type ContextVar struct {
	Name  string
	Type  string
	Value string
}

type UpdateCondition struct {
	Name    string
	Context []ContextVar
}

type PermissionUpdate struct {
	// The actor who owns the relation
	Subject string
	// The relation specifier
	Relation string
	// The target resource identifier
	Target    string
	Condition *UpdateCondition
}

type UpdatePermissionsRequest struct {
	Updates []PermissionUpdate
}
