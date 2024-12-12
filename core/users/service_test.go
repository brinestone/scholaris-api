package users_test

import (
	"context"
	"testing"

	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/core/users"
	dto "github.com/brinestone/scholaris/dto"
	"github.com/stretchr/testify/assert"
)

func mockEndpoints() {
	et.MockEndpoint(permissions.CheckPermissionInternal, func(ctx context.Context, req dto.InternalRelationCheckRequest) (*dto.RelationCheckResponse, error) {
		return &dto.RelationCheckResponse{
			Allowed: true,
		}, nil
	})
	et.MockEndpoint(permissions.SetPermissions, func(ctx context.Context, req dto.UpdatePermissionsRequest) error {
		return nil
	})
	// et.MockEndpoint(sAuth.VerifyCaptchaToken, func(ctx context.Context, req sAuth.VerifyCaptchaRequest) error {
	// 	return nil
	// })
	et.MockEndpoint(permissions.ListObjectsInternal, func(ctx context.Context, req dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTForm: {},
			},
		}, nil
	})
}

func TestMain(t *testing.M) {
	mockEndpoints()
	t.Run()
}

func makeUser() (res *dto.NewUserResponse, err error) {
	pass := gofakeit.Password(true, true, true, true, true, 20)
	person := gofakeit.Person()
	res, err = users.NewInternalUser(context.TODO(), dto.NewInternalUserRequest{
		FirstName:       person.FirstName,
		LastName:        person.LastName,
		Email:           person.Contact.Email,
		Dob:             gofakeit.PastDate().Format("2006/2/1"),
		Password:        pass,
		ConfirmPassword: pass,
		Phone:           person.Contact.Phone,
		Gender:          dto.Gender(person.Gender),
		CaptchaToken:    gofakeit.LoremIpsumSentence(40),
	})
	return
}

func TestNewUser(t *testing.T) {
	res, err := makeUser()

	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.Greater(t, res.UserId, uint64(0))
}

func TestFindUserByIdPublic(t *testing.T) {
	res, err := makeUser()
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	found, err := users.FindUserByIdPublic(context.TODO(), res.UserId)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, found)
	assert.Equal(t, res.UserId, found.Id)
}
