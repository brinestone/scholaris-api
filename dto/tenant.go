package dto

import "errors"

type NewTenantRequest struct {
	Name string `json:"name"`
}

func (n NewTenantRequest) Validate() error {
	if len(n.Name) == 0 {
		return errors.New("the name field is required")
	}
	return nil
}
