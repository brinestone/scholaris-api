package dto

type PermissionUpdate struct {
	User     string
	Relation string
	Object   string
}

type UpdatePermissionsRequest struct {
	Updates []PermissionUpdate
}
