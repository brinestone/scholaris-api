package tenants_test

import (
	"context"
	crypto "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/tenants"
	"github.com/stretchr/testify/assert"
)

func mockEndpoints() {
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
	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, req dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTForm: {},
			},
		}, nil
	})
	et.MockEndpoint(permissions.DeletePermissions, func(ctx context.Context, req dto.UpdatePermissionsRequest) error {
		return nil
	})
}

func randomString(len int) string {
	buf := make([]byte, len)
	io.ReadFull(crypto.Reader, buf)
	return hex.EncodeToString(buf)
}

func makeUser() (auth.UID, dto.AuthClaims) {
	uid := uint64(rand.Int63n(1000))
	userData := dto.AuthClaims{
		Email:      gofakeit.Person().Contact.Email,
		Avatar:     &gofakeit.Person().Image,
		FullName:   gofakeit.Name(),
		Provider:   gofakeit.RandomString(sAuth.ValidProviders),
		ExternalId: randomString(30),
		Account:    uint64(gofakeit.UintRange(1, 10000)),
		Sub:        uid,
	}
	return auth.UID(fmt.Sprintf("%d", uid)), userData
}

func makeTenant() (err error) {
	err = tenants.NewTenant(mainContext, dto.NewTenantRequest{
		Name:         gofakeit.Company(),
		CaptchaToken: randomString(30),
	})
	return
}

var mainContext context.Context

func TestMain(m *testing.M) {
	mockEndpoints()
	uid, data := makeUser()
	mainContext = auth.WithContext(context.TODO(), uid, &data)
	m.Run()
}

func TestFindSubscriptionPlans(t *testing.T) {
	res, err := tenants.FindSubscriptionPlans(context.TODO())
	if err != nil {
		t.Error(t, err)
		return
	}

	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Plans)
}

func TestNewTenant(t *testing.T) {
	err := makeTenant()
	assert.Nil(t, err)
}

func TestFindTenant(t *testing.T) {
	cnt := gofakeit.IntRange(1, 10)
	var err error
	for i := 0; i < cnt; i++ {
		err = makeTenant()
		if err != nil {
			t.Error(err)
			return
		}
	}

	t.Cleanup(mockEndpoints)
	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTTenant: {1},
			},
		}, nil
	})

	lookup, err := tenants.FindTenant(mainContext, 1)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, lookup)
	assert.Equal(t, uint64(1), lookup.Id)
}

func TestDeleteTenant(t *testing.T) {
	if err := makeTenant(); err != nil {
		t.Error(err)
		return
	}

	err := tenants.DeleteTenant(mainContext, 1)
	assert.Nil(t, err)
}

func TestLookup(t *testing.T) {
	t.Cleanup(mockEndpoints)
	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTTenant: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
		}, nil
	})
	if err := makeTenant(); err != nil {
		t.Error(err)
		return
	}

	res, err := tenants.Lookup(mainContext)

	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.LessOrEqual(t, len(res.Tenants), 100)
}
