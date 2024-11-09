package settings_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/settings"
	"github.com/stretchr/testify/assert"
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
			Relations: map[dto.PermissionType][]uint64{
				dto.PTForm: {},
			},
		}, nil
	})
	t.Run()
}

func makeUser() (auth.UID, sAuth.AuthClaims) {
	uid := uint64(rand.Int63n(1000))
	userData := sAuth.AuthClaims{
		Email:    gofakeit.Person().Contact.Email,
		Avatar:   gofakeit.Person().Image,
		FullName: gofakeit.Name(),
		Sub:      uid,
	}
	return auth.UID(fmt.Sprintf("%d", uid)), userData
}

func TestGetSettings(t *testing.T) {
	ownerId := uint64(gofakeit.UintRange(1, 5000))
	uid, data := makeUser()
	ctx := auth.WithContext(context.TODO(), uid, &data)

	res, err := settings.GetSettings(ctx, ownerId)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
}
