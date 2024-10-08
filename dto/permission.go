package dto

type PermissionType string

const (
	PTInstitution PermissionType = "institution"
	PTUser        PermissionType = "user"
	PTTenant      PermissionType = "tenant"
)

type ListRelationsResponse struct {
	// The valid relations
	Relations map[string][]string `json:"relations"`
}

type ListRelationsRequest struct {
	// The object claiming to own the relation.
	Subject string `query:"subject"`
	// The relation specifier
	Relation string `query:"relation"`
	// The target object
	Type string `query:"type"`
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
