package dto

type PermissionType string

func PermissionTypeFromString(s string) (PermissionType, bool) {
	switch s {
	case string(PTInstitution):
		return PTInstitution, true
	case string(PTUser):
		return PTUser, true
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
	Subject   string
	Relation  string
	Target    string
	Condition *UpdateCondition
}

type UpdatePermissionsRequest struct {
	Updates []PermissionUpdate
}
