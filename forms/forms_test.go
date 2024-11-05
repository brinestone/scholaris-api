package forms_test

import (
	"context"
	"encoding/hex"
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
	captchaToken := randomString(20)

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

	t.Run("TestDeleteForm", func(t *testing.T) {
		testDeleteForm(t, ctx, form.Id)
	})
}

func randomString(len int) string {
	buf := make([]byte, len)
	io.ReadFull(crypto.Reader, buf)
	return hex.EncodeToString(buf)
}

func testUpdateFormQuestionOptions(t *testing.T, ctx context.Context, form uint64, questionId uint64) {
	var options []dto.QuestionOption
	t.Run("Test_AddOptions", func(t *testing.T) {
		val := randomString(4)
		req := dto.UpdateFormQuestionOptionsRequest{
			Added: []dto.NewQuestionOption{
				{
					Caption:   randomString(5),
					Value:     &val,
					IsDefault: true,
				},
			},
		}

		res, err := forms.UpdateFormQuestionOptions(ctx, form, questionId, req)
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		var q *dto.FormQuestion
		for _, v := range res.Questions {
			if v.Id == questionId {
				q = &v
				break
			}
		}

		assert.NotNil(t, q)
		assert.Greater(t, len(q.Options), 0)
		assert.Equal(t, req.Added[0].Caption, q.Options[0].Caption)
		assert.Equal(t, req.Added[0].Value, q.Options[0].Value)
		assert.Equal(t, req.Added[0].IsDefault, q.Options[0].IsDefault)
		options = q.Options
	})

	t.Run("Test_UpdateOptions", func(t *testing.T) {
		img := randomString(20)
		newCaption := randomString(9)
		req := dto.UpdateFormQuestionOptionsRequest{
			Added: []dto.NewQuestionOption{
				{
					Caption:   newCaption,
					Value:     nil,
					IsDefault: false,
				},
			},
			Updates: []dto.FormQuestionOptionUpdate{
				{
					Id:        options[0].Id,
					Caption:   randomString(2),
					Value:     nil,
					Image:     &img,
					IsDefault: !options[0].IsDefault,
				},
			},
		}

		res, err := forms.UpdateFormQuestionOptions(ctx, form, questionId, req)
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		var q *dto.FormQuestion
		for _, v := range res.Questions {
			if v.Id == questionId {
				q = &v
				break
			}
		}

		assert.NotNil(t, q)
		assert.Greater(t, len(q.Options), len(options))
		assert.Len(t, q.Options, 2)
		assert.NotEqual(t, options[0].Caption, q.Options[0].Caption)
		assert.Nil(t, q.Options[0].Value)
		assert.False(t, q.Options[0].IsDefault)

		options = q.Options
	})

	t.Run("Test_RemoveWithConflict", func(t *testing.T) {
		img := randomString(20)
		newCaption := randomString(9)
		req := dto.UpdateFormQuestionOptionsRequest{
			Updates: []dto.FormQuestionOptionUpdate{
				{
					Id:        options[0].Id,
					Caption:   newCaption,
					Value:     nil,
					Image:     &img,
					IsDefault: !options[0].IsDefault,
				},
				{
					Id:        options[1].Id,
					Caption:   newCaption,
					Value:     nil,
					IsDefault: true,
				},
			},
			Removed: []uint64{options[0].Id, options[1].Id},
		}

		res, err := forms.UpdateFormQuestionOptions(ctx, form, questionId, req)
		assert.Nil(t, res)
		assert.Error(t, err)
	})

	t.Run("Test_RemoveWithoutConflict", func(t *testing.T) {
		req := dto.UpdateFormQuestionOptionsRequest{
			Removed: []uint64{
				options[0].Id,
			},
		}

		res, err := forms.UpdateFormQuestionOptions(ctx, form, questionId, req)
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.Less(t, len(res.Questions), len(options))
		assert.Equal(t, res.Questions[0].Id, uint64(1))
	})
}

func testUpdateQuestion(t *testing.T, ctx context.Context, form uint64, ref dto.FormQuestion) {
	updatedPrompt := randomString(20)
	req := dto.UpdateFormQuestionRequest{
		Prompt:        updatedPrompt,
		IsRequired:    ref.IsRequired,
		Type:          ref.Type,
		LayoutVariant: &ref.LayoutVariant,
	}

	res, err := forms.UpdateQuestion(ctx, form, ref.Id, req)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.Len(t, res.Questions, 1)
	assert.NotEqual(t, ref.Prompt, res.Questions[0].Prompt)
}

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

	assert.NotNil(t, res)
	assert.Len(t, res.Questions, 1)
	t.Run("Test_UpdateFormQuestionOptions", func(t *testing.T) {
		testUpdateFormQuestionOptions(t, ctx, refId, res.Questions[0].Id)
	})
	t.Run("Test_UpdateQuestion", func(t *testing.T) {
		testUpdateQuestion(t, ctx, refId, res.Questions[0])
	})
	t.Run("Test_DeleteQuestion", func(t *testing.T) {
		req := dto.DeleteQuestionsRequest{
			Questions: []uint64{res.Questions[0].Id},
		}
		res1, err := forms.DeleteFormQuestions(ctx, refId, req)
		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res1)
		assert.Less(t, len(res1.Questions), len(res.Questions))
	})
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

func testDeleteForm(t *testing.T, ctx context.Context, id uint64) {
	if err := forms.DeleteForm(ctx, id); err != nil {
		t.Error(err)
		return
	}

	res, err := forms.FindFormQuestions(ctx, id)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Empty(t, res.Questions)
}
