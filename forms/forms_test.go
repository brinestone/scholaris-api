package forms_test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"

	crypto "crypto/rand"

	"encore.dev/beta/auth"
	"encore.dev/et"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/forms"
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
	t.Run()
}

func TestForms(t *testing.T) {
	uid := uint64(rand.Int63n(400))
	captchaToken := make([]byte, 32)
	io.ReadFull(crypto.Reader, captchaToken)

	randomUserId := auth.UID(fmt.Sprintf("%d", uid))
	ctx := auth.WithContext(context.Background(), randomUserId, &sAuth.AuthClaims{
		Email:    "john@example.com",
		FullName: "John Doe",
		Sub:      uid,
	})
	ownerId := uint64(rand.Int63n(400))

	form, err := forms.NewForm(ctx, dto.NewFormInput{
		Title:           "Test Form",
		Description:     nil,
		BackgroundColor: nil,
		CaptchaToken:    string(captchaToken),
		Owner:           ownerId,
		OwnerType:       string(dto.PTInstitution),
	})
	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, form)
	assert.Greater(t, uint64(form.Id), uint64(0))

	t.Run("TestFindForms_Owned", func(t *testing.T) {
		testFindOwnedForms(t, ownerId, ctx, dto.PTInstitution)
	})

	t.Run("TestFindForms_Unowned", func(t *testing.T) {
		testFindUnOwnedForms(t, ownerId, ctx, dto.PTTenant, form.Id)
	})

	t.Run("TestUpdateForm", func(t *testing.T) {
		testUpdateForm(t, ctx, form)
	})

	t.Run("TestCreateQuestion", func(t *testing.T) {
		testCreateQuestions(t, ctx, form.Id)
	})
}

func randomString(len int) string {
	buf := make([]byte, len)
	io.ReadFull(crypto.Reader, buf)
	return string(buf)
}

// func randomStringPtr(len int) *string {
// 	s := randomString(len)
// 	return &s
// }

func testCreateQuestions(t *testing.T, ctx context.Context, refId uint64) {
	req := dto.UpdateFormQuestionRequest{
		Prompt:     randomString(10),
		IsRequired: true,
		Type:       dto.QTSingleline,
	}

	res, err := forms.CreateQuestion(ctx, refId, req)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Len(t, res.Questions, 1)
}

func testFindUnOwnedForms(t *testing.T, owner uint64, ctx context.Context, ownerType dto.PermissionType, refId uint64) {
	res, err := forms.FindForms(ctx, dto.GetFormsInput{
		Page:      0,
		Size:      10,
		Owner:     owner + 1,
		OwnerType: string(ownerType),
	})
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	for _, v := range res.Data {
		assert.NotEqual(t, refId, v.Id)
	}
}

func testFindOwnedForms(t *testing.T, owner uint64, ctx context.Context, ownerType dto.PermissionType) {
	res, err := forms.FindForms(ctx, dto.GetFormsInput{
		Page:      0,
		Size:      10,
		Owner:     owner,
		OwnerType: string(ownerType),
	})
	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, res)
	assert.Greater(t, res.Meta.Total, uint(0))
}

func testUpdateForm(t *testing.T, ctx context.Context, form *dto.FormConfig) {
	request := dto.UpdateFormRequest{
		Title:           fmt.Sprintf("%s %s", form.Title, "update"),
		Description:     form.Description,
		BackgroundColor: form.BackgroundColor,
		BackgroundImage: form.BackgroundImage,
		Image:           form.Image,
		MultiResponse:   !form.MultiResponse,
		Resubmission:    form.Resubmission,
		CaptchaToken:    "some captcha token",
	}

	res, err := forms.UpdateForm(ctx, form.Id, request)
	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, res)
	assert.Equal(t, !form.MultiResponse, res.MultiResponse)
	assert.NotEqual(t, form.Title, res.Title)
	assert.Greater(t, res.UpdateAt, form.UpdateAt)
}
