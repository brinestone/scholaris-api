// CRUD endpoints for institutions
package institutions

import (
	"context"

	"github.com/brinestone/scholaris/dto"
)

// Creates a new Institution
//
//encore:api auth method=POST path=/institutions tag:needs_captcha_ver tag:perm_can_create
func NewInstitution(ctx context.Context, req dto.NewInstitutionRequest) error {

	return nil
}

// Looks up a tenant's  institutions
//
//encore:api public method=GET path=/institutions
func LookupInstitutions(ctx context.Context, req dto.PageBasedPaginationParams) (*dto.PaginatedResponse[dto.InstitutionLookup], error) {
	return nil, nil
}
