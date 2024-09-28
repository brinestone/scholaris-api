// CRUD endpoints for institutions
package institutions

import "context"

// Creates a new institution
//
//encore:api auth method=POST path=/institutions tag:perm_can_create_institution
func New(ctx context.Context) error {
	return nil
}
