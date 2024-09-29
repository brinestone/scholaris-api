package dto

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
