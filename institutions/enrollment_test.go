package institutions_test

import (
	"context"
	"testing"

	"encore.dev/et"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
)

func TestMain(t *testing.M) {
	et.MockEndpoint(permissions.CheckPermission, func(ctx context.Context, req dto.RelationCheckRequest) (*dto.RelationCheckResponse, error) {
		return &dto.RelationCheckResponse{
			Allowed: true,
		}, nil
	})
	et.MockEndpoint(permissions.SetPermissions, func(ctx context.Context, req dto.UpdatePermissionsRequest) error {
		return nil
	})
	et.MockEndpoint(sAuth.VerifyCaptchaToken, func(ctx context.Context, req sAuth.VerifyCaptchaRequest) error {
		return nil
	})
	et.MockEndpoint(permissions.ListRelations, func(ctx context.Context, req dto.ListRelationsRequest) (*dto.ListRelationsResponse, error) {
		return &dto.ListRelationsResponse{
			Relations: map[dto.PermissionType][]string{
				dto.PTForm: {},
			},
		}, nil
	})
	t.Run()
}

func TestNewInstitution(t *testing.T) {

}

func TestLookupInstitutions(t *testing.T) {

}
