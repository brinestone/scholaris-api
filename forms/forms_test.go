package forms_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	crypto "crypto/rand"

	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/forms"
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
	et.MockEndpoint(permissions.ListRelations, func(ctx context.Context, req dto.ListRelationsRequest) (*dto.ListRelationsResponse, error) {
		return &dto.ListRelationsResponse{
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

type makeFormOptionName string

const (
	questionCount makeFormOptionName = "question-count"
)

type makeFormOption struct {
	name  makeFormOptionName
	value any
}

func withOption(name makeFormOptionName, value any) makeFormOption {
	return makeFormOption{
		name:  name,
		value: value,
	}
}

func makeFormQuestions(ctx context.Context, form uint64, count uint) error {
	var gErr error
	for i := uint(0); i < count; i++ {
		questionRequest := dto.UpdateFormQuestionRequest{
			Prompt:     gofakeit.LoremIpsumSentence(gofakeit.IntRange(3, 15)),
			IsRequired: gofakeit.Bool(),
			Type:       gofakeit.RandomString([]string{string(dto.QTDate), string(dto.QTEmail), string(dto.QTFile), string(dto.QTGeoPoint), string(dto.QTMCQ), string(dto.QTMultiline), string(dto.QTSingleChoice), string(dto.QTSingleline), string(dto.QTTel)}),
		}
		q, err := forms.CreateQuestion(ctx, form, questionRequest)
		if err != nil {
			gErr = err
			break
		}

		if questionRequest.Type == dto.QTMCQ || questionRequest.Type == dto.QTSingleChoice {
			optionRequest := dto.UpdateFormQuestionOptionsRequest{}
			cnt := gofakeit.IntRange(1, 30)
			for j := 0; j < cnt; j++ {
				var img, val *string

				if gofakeit.Bool() {
					tmp := gofakeit.ImageURL(48, 48)
					img = &tmp
				}

				if questionRequest.IsRequired || gofakeit.Bool() {
					tmp := strings.ToLower(gofakeit.LoremIpsumWord())
					val = &tmp
				}

				op := dto.NewQuestionOption{
					Caption:   gofakeit.LoremIpsumSentence(gofakeit.IntRange(2, 5)),
					IsDefault: gofakeit.Bool(),
					Value:     val,
					Image:     img,
				}

				optionRequest.Added = append(optionRequest.Added, op)
			}
			_, err = forms.UpdateFormQuestionOptions(ctx, form, q.Questions[len(q.Questions)-1].Id, optionRequest)
			if err != nil {
				gErr = err
				break
			}
		}
	}
	return gErr
}

func makeForm(ctx context.Context, options ...makeFormOption) (owner uint64, _ *dto.FormConfig, _ error) {
	ownerId := uint64(rand.Int63n(10000))
	desc := gofakeit.LoremIpsumParagraph(1, rand.Intn(3), rand.Intn(10), "\n")
	var image, bg, bgImg *string
	if gofakeit.Bool() {
		tmp := gofakeit.ImageURL(100, 100)
		image = &tmp
	}
	if gofakeit.Bool() {
		tmp := gofakeit.HexColor()
		bg = &tmp
	}
	if gofakeit.Bool() {
		tmp := gofakeit.ImageURL(640, 640)
		bgImg = &tmp
	}

	var tagsCount = gofakeit.IntRange(0, 10)
	var tags []string

	for i := 0; i < tagsCount; i++ {
		tags = append(tags, gofakeit.LoremIpsumWord())
	}

	window := time.Hour * time.Duration(gofakeit.IntRange(1, 3000))
	responseStart := gofakeit.FutureDate()
	res, err := forms.NewForm(ctx, dto.NewFormInput{
		Title:           gofakeit.LoremIpsumSentence(gofakeit.IntRange(1, 10)),
		Description:     &desc,
		CaptchaToken:    randomString(20),
		Owner:           ownerId,
		OwnerType:       gofakeit.RandomString([]string{string(dto.PTInstitution), string(dto.PTTenant)}),
		ResponseStart:   &responseStart,
		ResponseWindow:  &window,
		MultiResponse:   gofakeit.Bool(),
		Resubmission:    gofakeit.Bool(),
		Image:           image,
		BackgroundColor: bg,
		Tags:            tags,
		BackgroundImage: bgImg,
	})
	if err != nil {
		return 0, nil, err
	}

	if len(options) > 0 {
		for _, v := range options {
			switch makeFormOptionName(v.name) {
			case questionCount:
				cnt, ok := v.value.(uint)
				if !ok {
					cnt = gofakeit.UintRange(1, 20)
				}
				makeFormQuestions(ctx, res.Id, cnt)
			}
		}
	}

	form, err := forms.GetFormInfo(ctx, res.Id)
	if err != nil {
		return ownerId, nil, err
	}

	return ownerId, form, err
}

func TestForms(t *testing.T) {
	user, userData := makeUser()
	ctx := auth.WithContext(context.TODO(), user, &userData)

	ownerId, form, err := makeForm(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, form)
	assert.Greater(t, uint64(form.Id), uint64(0))

	t.Run("Test_GetFormInfo", func(t *testing.T) {
		t.Run("Test_GetFormInfo_Exists", func(t *testing.T) {
			testFindExistingFormInfo(t, ctx, form.Id)
		})
		t.Run("Test_GetFormInfo_NoExists", func(t *testing.T) {
			testFindNonExistingFormInfo(t, ctx)
		})
	})

	t.Run("TestFindForms_Owned", func(t *testing.T) {
		et.MockEndpoint(permissions.ListRelations, func(ctx context.Context, p dto.ListRelationsRequest) (ans *dto.ListRelationsResponse, err error) {
			ans = new(dto.ListRelationsResponse)
			ans.Relations = make(map[dto.PermissionType][]uint64)
			ans.Relations[dto.PermissionType(p.Type)] = []uint64{form.Id}
			return
		})
		testFindOwnedForms(t, ownerId, ctx, dto.PTInstitution)
		mockEndpoints()
	})

	t.Run("TestFindForms_Unowned", func(t *testing.T) {
		testFindUnOwnedForms(t, ownerId, ctx, dto.PTTenant, form.Id)
	})

	var groupId uint64
	t.Run("Test_Grouping", func(t *testing.T) {
		testFormQuestionGrouping(t, ctx, form.Id, &groupId)
	})

	t.Run("TestUpdateForm", func(t *testing.T) {
		testUpdateForm(t, ctx, form)
	})

	t.Run("TestCreateQuestion", func(t *testing.T) {
		testCreateQuestions(t, ctx, form.Id, groupId)
	})

	t.Run("Test_GroupDeletion", func(t *testing.T) {
		res, err := forms.DeleteQuestionGroup(ctx, form.Id, dto.DeleteFormQuestionGroupsRequest{
			Ids: []uint64{
				groupId,
			},
		})

		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.Empty(t, res.Groups)
		assert.Empty(t, res.Questions)
	})

	t.Run("TestToggleFormStatus", func(t *testing.T) {
		testFormToggle(t, ctx, form)
	})

	t.Run("TestDeleteForm", func(t *testing.T) {
		testDeleteForm(t, ctx, form.Id)
	})
}

func testFormToggle(t *testing.T, ctx context.Context, form *dto.FormConfig) {
	update, err := forms.ToggleFormStatus(ctx, form.Id)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, update)
	assert.NotEqual(t, form.Status, update.Status)
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

func testCreateQuestions(t *testing.T, ctx context.Context, refId, group uint64) {
	req := dto.UpdateFormQuestionRequest{
		Prompt:     randomString(10),
		IsRequired: true,
		Type:       dto.QTSingleline,
		Group:      group,
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
	res, err := forms.FindForms(ctx, dto.FindFormsRequest{
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
	for _, v := range res.Forms {
		assert.NotEqual(t, refId, v.Id)
	}
}

func testFindOwnedForms(t *testing.T, owner uint64, ctx context.Context, ownerType dto.PermissionType) {
	res, err := forms.FindForms(ctx, dto.FindFormsRequest{
		Page:      0,
		Size:      10,
		Owner:     owner,
		OwnerType: string(ownerType),
	})
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	// assert.NotEmpty(t, res.Forms)
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
		return
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

func testFindExistingFormInfo(t *testing.T, ctx context.Context, id uint64) {
	res, err := forms.GetFormInfo(ctx, id)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.Equal(t, id, res.Id)
}

func testFindNonExistingFormInfo(t *testing.T, ctx context.Context) {
	randomId := uint64(rand.Int63n(1000))
	res, err := forms.GetFormInfo(ctx, randomId)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func testFormQuestionGrouping(t *testing.T, ctx context.Context, formId uint64, ref *uint64) {
	var group dto.FormQuestionGroup
	t.Run("Test_NewGroup", func(t *testing.T) {
		label := randomString(20)
		res, err := forms.CreateQuestionGroup(ctx, formId, dto.UpdateFormQuestionGroupRequest{
			Label: &label,
		})

		if err != nil {
			t.Error(err)
			return
		}

		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Groups)
		group = res.Groups[0]
		*ref = group.Id
	})

	t.Run("Test_UpdateGroup", func(t *testing.T) {
		updatedLabel := *group.Label + " updated"
		res, err := forms.UpdateQuestionGroup(ctx, formId, group.Id, dto.UpdateFormQuestionGroupRequest{
			Label:       &updatedLabel,
			Description: group.Description,
			Image:       group.Image,
		})

		if err != nil {
			t.Error(err)
		}

		assert.NotNil(t, res)
		assert.NotEqual(t, updatedLabel, res.Groups[0].Label)
	})
}

func TestResponses(t *testing.T) {
	uid, userData := makeUser()
	ctx := auth.WithContext(context.TODO(), uid, &userData)
	_, form, err := makeForm(ctx, withOption(questionCount, gofakeit.UintRange(1, 15)))
	if err != nil {
		t.Error(err)
		return
	}
	otherUid, otherUserData := makeUser()
	et.OverrideAuthInfo(otherUid, &otherUserData)

	passed := t.Run("Status=draft", func(t *testing.T) {
		testCreateFormResponseWhileInDraft(t, ctx, form.Id)
	})

	if passed {
		et.OverrideAuthInfo(uid, &userData)
		form, err = forms.ToggleFormStatus(ctx, form.Id)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("Status=published", func(t *testing.T) {
			et.OverrideAuthInfo(otherUid, &otherUserData)
			testCreateFormResponseWhilePublished(t, ctx, form.Id)
		})

		req := dto.UpdateFormRequest{
			MultiResponse:   true,
			Title:           form.Title,
			Description:     form.Description,
			BackgroundColor: form.BackgroundColor,
			BackgroundImage: form.BackgroundImage,
			Image:           form.Image,
			Resubmission:    form.Resubmission,
			CaptchaToken:    randomString(20),
			Deadline:        form.Deadline,
			MaxResponses:    form.MaxResponses,
			MaxSubmissions:  form.MaxSubmissions,
		}
		form, err := forms.UpdateForm(ctx, form.Id, req)

		if err != nil {
			t.Error(err)
			return
		}

		if form.MultiResponse && form.MaxResponses == nil {
			t.Run("Multiresponse&unlimited", func(t *testing.T) {
				testCreateMultiResponseUnlimited(t, ctx, form.Id)
			})
		}

		if form.MaxResponses == nil {
			tmp := gofakeit.UintRange(5, 20)
			req.MaxResponses = &tmp
			form, err = forms.UpdateForm(ctx, form.Id, req)
			if err != nil {
				t.Error(err)
				return
			}
		}

		if form.MultiResponse && form.MaxResponses != nil && *form.MaxResponses > 0 {
			t.Run("MultiResponse&limited", func(t *testing.T) {
				testCreateMultiResponseLimited(t, ctx, form.Id, *form.MaxResponses)
			})
		}

		t.Run("GetAllResponses", func(t *testing.T) {
			testGetUserResponses(t, ctx, form.Id)
		})
		t.Run("GetSingleResponse", func(t *testing.T) {
			testGetUserResponse(t, ctx, form.Id)
		})

		t.Run("UpdateAnswers", func(t *testing.T) {
			testUpdateResponseAnswers(t, ctx, form.Id)
		})

		t.Run("SubmitResponse", func(t *testing.T) {
			testResponseSubmission(t, ctx, form.Id)
		})
	}
}

func testResponseSubmission(t *testing.T, ctx context.Context, form uint64) {
	responses, err := forms.GetUserResponses(ctx, form)
	if err != nil {
		t.Error(err)
		return
	}

	chosenIndex := gofakeit.IntRange(0, len(responses.Responses)-1)
	response := responses.Responses[chosenIndex]
	res, err := forms.SubmitResponse(ctx, form, response.Id)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.NotNil(t, res.SubmittedAt)
}

func testUpdateResponseAnswers(t *testing.T, ctx context.Context, form uint64) {
	questions, err := forms.FindFormQuestions(ctx, form)
	if err != nil {
		t.Error(err)
		return
	}

	req := dto.UpdateUserAnswersRequest{}
	questionMap := make(map[uint64]dto.FormQuestion)
	for _, v := range questions.Questions {
		questionMap[v.Id] = v
		var ans = dto.FormAnswerUpdate{
			Question: v.Id,
		}

		switch v.Type {
		case dto.QTSingleline:
			tmp := gofakeit.LoremIpsumSentence(gofakeit.IntRange(1, 5))
			ans.Value = &tmp
		case dto.QTDate:
			tmp := gofakeit.Date().String()
			ans.Value = &tmp
		case dto.QTEmail:
			tmp := gofakeit.Email()
			ans.Value = &tmp
		case dto.QTFile:
			tmp := gofakeit.URL()
			ans.Value = &tmp
		case dto.QTGeoPoint:
			tmp := fmt.Sprintf("%f,%f", gofakeit.Longitude(), gofakeit.Latitude())
			ans.Value = &tmp
		case dto.QTMultiline:
			tmp := gofakeit.LoremIpsumParagraph(1, 2, 20, "\n")
			ans.Value = &tmp
		case dto.QTTel:
			tmp := gofakeit.Phone()
			ans.Value = &tmp
		case dto.QTMCQ, dto.QTSingleChoice:
			cnt := 1
			if v.Type == dto.QTMCQ {
				cnt = gofakeit.IntRange(1, 10)
			}

			chosenOptions := make([]string, 0)
			for i := 0; i < cnt; i++ {
				index := gofakeit.IntRange(0, len(v.Options)-1)
				if v.Options[index].Value == nil {
					i--
					continue
				}
				chosenOptions = append(chosenOptions, *v.Options[index].Value)
			}

			tmp := strings.Join(chosenOptions, ",")
			ans.Value = &tmp
		default:
			ans.Value = nil
		}

		req.Updated = append(req.Updated, ans)
	}

	res, err := forms.UpdateResponseAnswers(ctx, form, 1, &req)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Answers)
}

func testGetUserResponse(t *testing.T, ctx context.Context, form uint64) {
	r, err := forms.GetUserResponse(ctx, form, 1)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, r)
	assert.Equal(t, uint64(1), r.Id)
}

func testGetUserResponses(t *testing.T, ctx context.Context, form uint64) {
	r, err := forms.GetUserResponses(ctx, form)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, r)
	assert.NotEmpty(t, r.Responses)
}

func testCreateMultiResponseLimited(t *testing.T, ctx context.Context, form uint64, max uint) {
	var r *dto.UserFormResponses
	var err error
	var created uint

	for i := uint(0); i < max+10; i++ {
		r, err = forms.CreateFormResponse(ctx, form)
		if err == nil {
			created++
		}
	}

	if r != nil {
		assert.LessOrEqual(t, len(r.Responses), max)
	} else {
		assert.NotNil(t, err)
	}
}

func testCreateMultiResponseUnlimited(t *testing.T, ctx context.Context, form uint64) {
	var r *dto.UserFormResponses
	var cnt = gofakeit.IntRange(5, 20)

	for i := 0; i < cnt; i++ {
		var err error
		r, err = forms.CreateFormResponse(ctx, form)
		if err != nil {
			t.Error(err)
			return
		}
	}

	assert.NotNil(t, r)
	assert.GreaterOrEqual(t, len(r.Responses), 6)
}

func testCreateFormResponseWhilePublished(t *testing.T, ctx context.Context, form uint64) {
	responses, err := forms.CreateFormResponse(ctx, form)
	assert.NotNil(t, responses)
	assert.Nil(t, err)
	assert.NotEmpty(t, responses.Responses)
}

func testCreateFormResponseWhileInDraft(t *testing.T, ctx context.Context, form uint64) {
	responses, err := forms.CreateFormResponse(ctx, form)

	assert.Nil(t, responses)
	assert.NotNil(t, err)
}
