// CRUD endpoints for forms
package forms

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
)

// Submits a user's response
//
//encore:api auth method=PATCH path=/forms/:form/responses/:response/submit tag:user_owns_response tag:user_can_submit_response
func SubmitResponse(ctx context.Context, form, response uint64) (*dto.UserFormResponse, error) {
	sub, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := submitUserResponse(ctx, tx, form, uid, response); err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		} else {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}
	}
	tx.Commit()

	defer func() {
		FormSubmissions.Publish(ctx, ResponseSubmitted{
			Form:      form,
			Response:  response,
			Timestamp: time.Now(),
		})
	}()

	r, err := findUserResponseById(ctx, form, uid, response)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &util.ErrNotFound
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	return &responsesToDto(r)[0], nil
}

// Updates a user's answers
//
//encore:api auth method=PATCH path=/forms/:form/responses/:response/answers tag:user tag:user_owns_response
func UpdateResponseAnswers(ctx context.Context, form, response uint64, req *dto.UpdateUserAnswersRequest) (*dto.UserFormResponse, error) {
	sub, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if len(req.Updated) > 0 {
		if err := updateUserResponseAnswers(ctx, tx, form, response, req.Updated...); err != nil {
			if errs.Convert(err) == nil {
				return nil, err
			} else {
				rlog.Error(util.MsgDbAccessError, "msg", err.Error())
				return nil, &util.ErrUnknown
			}
		}
	}

	if len(req.Removed) > 0 {
		if err := deleteResponseAnswers(ctx, tx, form, response, req.Removed...); err != nil {
			if errs.Convert(err) == nil {
				return nil, err
			} else {
				rlog.Error(util.MsgDbAccessError, "msg", err.Error())
				return nil, &util.ErrUnknown
			}
		}
	}
	tx.Commit()

	r, err := findUserResponseById(ctx, form, uid, response)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &util.ErrNotFound
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	return &responsesToDto(r)[0], nil
}

// Gets a user's response
//
//encore:api auth method=GET path=/forms/:form/responses/:response
func GetUserResponse(ctx context.Context, form, response uint64) (*dto.UserFormResponse, error) {
	sub, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	r, err := findUserResponseById(ctx, form, uid, response)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &util.ErrNotFound
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	return &responsesToDto(r)[0], nil
}

// Gets a user's form responses
//
//encore:api auth method=GET path=/forms/:form/responses
func GetUserResponses(ctx context.Context, form uint64) (*dto.UserFormResponses, error) {
	sub, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	responses, err := findUserResponses(ctx, uid, form)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	return &dto.UserFormResponses{
		Responses: responsesToDto(responses...),
	}, nil
}

// Creates a response for a form
//
//encore:api auth method=POST path=/forms/:form/responses/new tag:user_can_respond_to_form
func CreateFormResponse(ctx context.Context, form uint64) (*dto.UserFormResponses, error) {
	sub, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := createUserFormResponse(ctx, tx, form, uid); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	responses, err := findUserResponses(ctx, uid, form)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	return &dto.UserFormResponses{
		Responses: responsesToDto(responses...),
	}, nil
}

// Deletes a form's question group
//
//encore:api auth method=DELETE path=/forms/:form/groups tag:user_is_form_editor
func DeleteQuestionGroup(ctx context.Context, form uint64, req dto.DeleteFormQuestionGroupsRequest) (*dto.GetFormQuestionsResponse, error) {

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := deleteFormQuestionGroups(ctx, tx, form, req.Ids...); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Updates a form's question group
//
//encore:api auth method=PATCH path=/forms/:form/groups/:group tag:user_is_form_editor
func UpdateQuestionGroup(ctx context.Context, form, group uint64, req dto.UpdateFormQuestionGroupRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := updateFormQuestionGroup(ctx, tx, form, group, req); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Creates a form's question group
//
//encore:api auth method=POST path=/forms/:form/groups tag:user_is_form_editor
func CreateQuestionGroup(ctx context.Context, form uint64, req dto.UpdateFormQuestionGroupRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := createQuestionsGroup(ctx, tx, form, req); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	tx.Commit()
	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Gets a form's info
//
//encore:api public method=GET path=/forms/:form
func GetFormInfo(ctx context.Context, form uint64) (*dto.FormConfig, error) {
	cfg, err := formCache.Get(ctx, form)
	if err != nil {
		if !errors.Is(err, cache.Miss) {
			rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
		}

		_cfg, err := findFormFromDb(ctx, form)
		if err != nil {
			if errs.Convert(err) == nil {
				return nil, err
			} else if errors.Is(err, sqldb.ErrNoRows) {
				return nil, &util.ErrNotFound
			} else {
				rlog.Error(util.MsgDbAccessError, "msg", err.Error())
				return nil, &util.ErrUnknown
			}
		}

		ans := formsToDto(_cfg)
		if err := formCache.Set(ctx, form, ans[0]); err != nil {
			rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
		}

		cfg = ans[0]
	}

	uid, authed := auth.UserID()
	if !authed && cfg.Status == dto.FSDraft {
		rlog.Warn("draft form access attempt", "form", form)
		return nil, &util.ErrForbidden
	}

	perm, err := permissions.CheckPermission(ctx, dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermCanView,
		Target:   dto.IdentifierString(dto.PTForm, form),
	})

	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err.Error())
		return nil, &util.ErrUnknown
	} else if !perm.Allowed {
		return nil, &util.ErrForbidden
	}

	return &cfg, nil
}

// Toggles a form's status
//
//encore:api auth method=PUT path=/forms/:form/toggle tag:user_is_form_editor
func ToggleFormStatus(ctx context.Context, form uint64) (*dto.FormConfig, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := toggleFormStatus(ctx, tx, form); err != nil {
		tx.Rollback()
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, &util.ErrNotFound
		} else {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}
	}
	tx.Commit()

	f, err := findFormFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := formsToDto(f)
	if err := formCache.Set(ctx, form, ans[0]); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	if f.Status == "published" {
		PublishedForms.Publish(ctx, FormPublished{
			Id:        f.Id,
			Timestamp: f.UpdatedAt,
		})
	}
	return &ans[0], nil
}

// Deletes a form
//
//encore:api auth method=DELETE path=/forms/:form tag:user_is_form_editor
func DeleteForm(ctx context.Context, form uint64) error {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	if err := deleteForm(ctx, tx, form); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}
	tx.Commit()

	DeletedForms.Publish(ctx, FormDeleted{
		Id:        form,
		Timestamp: time.Now(),
	})

	return nil
}

// Deletes a form's questions
//
//encore:api auth method=DELETE path=/forms/:form/questions tag:user_is_form_editor
func DeleteFormQuestions(ctx context.Context, form uint64, req dto.DeleteQuestionsRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := deleteFormQuestions(ctx, tx, form, req.Questions...); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Updates a form question's options
//
//encore:api auth method=PATCH path=/forms/:form/questions/:question/options tag:user_is_form_editor
func UpdateFormQuestionOptions(ctx context.Context, form uint64, question uint64, req dto.UpdateFormQuestionOptionsRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if len(req.Updates) > 0 {
		if err := updateFormQuestionOptions(ctx, tx, form, question, req.Updates...); err != nil {
			tx.Rollback()
			rlog.Error(err.Error())
			return nil, &util.ErrUnknown
		}
	}

	if len(req.Added) > 0 {
		if err := createFormQuestionOptions(ctx, tx, form, question, req.Added...); err != nil {
			tx.Rollback()
			rlog.Error(err.Error())
			return nil, &util.ErrUnknown
		}
	}

	if len(req.Removed) > 0 {
		if err := deleteFormQuestionOptions(ctx, tx, form, question, req.Removed...); err != nil {
			tx.Rollback()
			rlog.Error(err.Error())
			return nil, &util.ErrUnknown
		}
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Updates a form question
//
//encore:api auth method=PATCH path=/forms/:form/questions/:question tag:user_is_form_editor
func UpdateQuestion(ctx context.Context, form uint64, question uint64, req dto.UpdateFormQuestionRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := updateFormQuestion(ctx, tx, form, question, req); err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		}
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &ans, nil
}

// Add a question to a form
//
//encore:api auth method=POST path=/forms/:form/question tag:user_is_form_editor
func CreateQuestion(ctx context.Context, form uint64, req dto.UpdateFormQuestionRequest) (*dto.GetFormQuestionsResponse, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := createFormQuestion(ctx, tx, form, req); err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		}
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	questions, groups, err := findFormQuestionsFromDb(ctx, form)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}
	if err := questionsCache.Set(ctx, questionsCacheKey(form), ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return &ans, nil
}

// Update a form
//
//encore:api auth method=PUT path=/forms/:id tag:user_is_form_editor
func UpdateForm(ctx context.Context, id uint64, req dto.UpdateFormRequest) (*dto.FormConfig, error) {
	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := updateForm(ctx, tx, id, req); err != nil {
		rlog.Error(err.Error())
		tx.Rollback()
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	form, err := findFormFromDb(ctx, id)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}

	ans := formsToDto(form)
	if err := formCache.Set(ctx, id, ans[0]); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return &ans[0], nil
}

// Gets an owner's form data
//
//encore:api public method=GET path=/forms/:id/questions
func FindFormQuestions(ctx context.Context, id uint64) (*dto.GetFormQuestionsResponse, error) {
	cacheKey := questionsCacheKey(id)
	response, err := questionsCache.Get(ctx, cacheKey)
	if errors.Is(err, cache.Miss) {
		questions, groups, err := findFormQuestionsFromDb(ctx, id)
		if err != nil {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}

		response = dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions), Groups: formQuestionGroupsToDto(groups)}

		if len(response.Questions) > 0 {
			if err := questionsCache.Set(ctx, cacheKey, response); err != nil {
				rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
			}
		}
	} else if err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}
	return &response, nil
}

// Gets forms of an owner
//
//encore:api public method=GET path=/forms
func FindForms(ctx context.Context, params dto.FindFormsRequest) (*dto.GetFormsResponse, error) {
	ownerType, _ := dto.PermissionTypeFromString(params.OwnerType)
	var overrides []uint64
	uid, authed := auth.UserID()
	if authed {
		res, err := permissions.ListRelations(ctx, dto.ListRelationsRequest{
			Actor:    dto.IdentifierString(dto.PTUser, uid),
			Relation: models.PermEditor,
			Type:     string(dto.PTForm),
		})
		if err != nil {
			rlog.Error(util.MsgCallError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}
		overrides = res.Relations[dto.PTForm]
	}

	res, err := findFormsFromCache(ctx, int(params.Page), int(params.Size), ownerType, params.Owner)
	if errors.Is(err, cache.Miss) {
		formsFromDb, err := findFormsFromDb(ctx, params.Page, params.Size, params.Owner, params.OwnerType, overrides)
		rlog.Debug("here", "overrides", overrides)
		if err != nil {
			return nil, err
		}

		var forms = formsToDto(formsFromDb...)

		response := &dto.GetFormsResponse{
			Forms: forms,
		}

		if len(formsFromDb) > 0 {
			key := formsCacheKey(int(params.Page), int(params.Size), ownerType, params.Owner)
			if err := formsCache.Set(ctx, key, *response); err != nil {
				rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
			}
		}
		res = response
	} else if err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return res, nil
}

// Creates a new form
//
//encore:api auth method=POST path=/forms tag:needs_captcha_ver
func NewForm(ctx context.Context, req dto.NewFormInput) (response dto.NewFormResponse, err error) {
	uid, _ := auth.UserID()
	pt, _ := dto.PermissionTypeFromString(req.OwnerType)
	permission, err := permissions.CheckPermission(ctx, dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermCanCreateForms,
		Target:   dto.IdentifierString(pt, req.Owner),
	})
	if err != nil || !permission.Allowed {
		if err != nil && errs.Convert(err) != nil {
			rlog.Warn(util.MsgCallError, "msg", err.Error())
		}
		err = &util.ErrForbidden
		return
	}

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	response.Id, err = createForm(ctx, tx, req.Owner, req)
	if err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return
		} else {
			rlog.Error(err.Error())
			err = &util.ErrUnknown
		}
	}

	update := dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Actor:    dto.IdentifierString(dto.PermissionType(req.OwnerType), req.Owner),
				Relation: models.PermOwner,
				Target:   dto.IdentifierString(dto.PTForm, response.Id),
			}, {
				Actor:    dto.IdentifierString(dto.PTUser, uid),
				Relation: models.PermEditor,
				Target:   dto.IdentifierString(dto.PTForm, response.Id),
			},
		},
	}

	if err := permissions.SetPermissions(ctx, update); err != nil {
		rlog.Error(err.Error())
		tx.Rollback()
		err = &util.ErrUnknown
	}
	tx.Commit()

	NewForms.Publish(ctx, FormEvent{
		Id:        response.Id,
		Timestamp: time.Now(),
	})
	return
}

func createForm(ctx context.Context, tx *sqldb.Tx, owner uint64, input dto.NewFormInput) (id uint64, err error) {
	query := `
		INSERT INTO
			forms(
				title,
				description,
				meta_background,
				meta_bg_img,
				meta_img,
				owner,
				multi_response,
				response_resubmission,
				deadline,
				max_responses,
				max_submissions,
				owner_type,
				tags
			)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id;
	`

	var description, bg, bgImg, img *string
	var deadline *time.Time
	var maxResponses, maxSubmissions *uint

	if input.MaxSubmissions > 0 {
		maxSubmissions = &input.MaxSubmissions
	}
	if input.MaxResponses > 0 {
		maxResponses = &input.MaxResponses
	}
	if input.Description != nil {
		description = input.Description
	}
	if input.BackgroundColor != nil {
		bg = input.BackgroundColor
	}
	if input.BackgroundImage != nil {
		bgImg = input.BackgroundImage
	}
	if input.Image != nil {
		img = input.Image
	}
	if input.Deadline != nil {
		deadline = input.Deadline
	}

	if err = tx.QueryRow(ctx, query, input.Title, description, bg, bgImg, img, owner, input.MultiResponse, input.Resubmission, deadline, maxResponses, maxSubmissions, input.OwnerType, pq.Array(input.Tags)).Scan(&id); err != nil {
		return
	}

	return
}

func findFormQuestionsFromDb(ctx context.Context, id uint64) ([]*models.FormQuestion, []*models.FormQuestionGroup, error) {
	questionsQuery := `
		SELECT
			*
		FROM
			vw_AllQuestionOptions a
		WHERE
			a.form = $1
		;
	`

	rows, err := formsDb.Query(ctx, questionsQuery, id)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ans []*models.FormQuestion = make([]*models.FormQuestion, 0)
	for rows.Next() {
		var q = new(models.FormQuestion)
		var optionsJson string
		if err := rows.Scan(&q.Id, &q.Prompt, &q.IsRequired, &q.Type, &q.LayoutVariant, &optionsJson, &q.Group, &q.Form); err != nil {
			return nil, nil, err
		}

		if err := json.Unmarshal([]byte(optionsJson), &q.Options); err != nil {
			return nil, nil, err
		}
		ans = append(ans, q)
	}
	var groups []*models.FormQuestionGroup = make([]*models.FormQuestionGroup, 0)

	groupsQuery := `
		SELECT
			id,
			form,
			label,
			description,
			image
		FROM
			form_question_groups
		WHERE
			form=$1;
	`

	groupRows, err := formsDb.Query(ctx, groupsQuery, id)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var group = new(models.FormQuestionGroup)
		if err := groupRows.Scan(&group.Id, &group.Form, &group.Label, &group.Description, &group.Image); err != nil {
			return nil, nil, err
		}
		groups = append(groups, group)
	}

	return ans, groups, nil
}

func findFormFromDb(ctx context.Context, id uint64) (form *models.Form, err error) {
	query := `SELECT * FROM vw_AllForms WHERE id=$1;`
	form = new(models.Form)
	var questionIdsJson, groupIdsJson, tagsJson string
	if err = formsDb.QueryRow(ctx, query, id).Scan(&form.Id, &form.Title, &form.Description, &form.BackgroundColor, &form.BackgroundImage, &form.Image, &form.CreatedAt, &form.UpdatedAt, &form.Owner, &form.OwnerType, &form.MultiResponse, &form.Resubmission, &form.Status, &form.Deadline, &questionIdsJson, &groupIdsJson, &form.ResponseCount, &form.SubmissionCount, &tagsJson, &form.MaxResponses); err != nil {
		form = nil
		return
	}

	if err = json.Unmarshal([]byte(questionIdsJson), &form.QuestionIds); err != nil {
		return
	}

	if err = json.Unmarshal([]byte(groupIdsJson), &form.GroupIds); err != nil {
		return
	}

	if err = json.Unmarshal([]byte(tagsJson), &form.Tags); err != nil {
		return
	}
	return
}

func findFormsFromDb(ctx context.Context, page, size int, owner uint64, ownerType string, overrides []uint64) (ans []*models.Form, err error) {
	query := `SELECT * FROM func_find_forms($1,$2,$3,$4,$5);`
	rows, err := formsDb.Query(ctx, query, owner, ownerType, pq.Array(overrides), page, size)
	if err != nil {
		return
	}
	defer rows.Close()

	var cnt = 0
	for rows.Next() {
		cnt++
		var form = new(models.Form)
		var questionIdsJson, groupIdsJson, tagsJson string
		if err = formsDb.QueryRow(ctx, query, owner, ownerType, pq.Array(overrides), page, size).Scan(&form.Id, &form.Title, &form.Description, &form.BackgroundColor, &form.BackgroundImage, &form.Image, &form.CreatedAt, &form.UpdatedAt, &form.Owner, &form.OwnerType, &form.MultiResponse, &form.Resubmission, &form.Status, &form.Deadline, &questionIdsJson, &groupIdsJson, &form.ResponseCount, &form.SubmissionCount, &tagsJson, &form.MaxResponses); err != nil {
			form = nil
			return
		}

		if err = json.Unmarshal([]byte(questionIdsJson), &form.QuestionIds); err != nil {
			return
		}

		if err = json.Unmarshal([]byte(groupIdsJson), &form.GroupIds); err != nil {
			return
		}

		if err = json.Unmarshal([]byte(tagsJson), &form.Tags); err != nil {
			return
		}
		ans = append(ans, form)
	}
	rlog.Debug("test", "ans", ans, "cnt", cnt)
	return
}

func questionsCacheKey(form uint64) string {
	uid, _ := auth.UserID()
	temp := fmt.Sprintf("%d%s", form, uid)
	sum := md5.Sum([]byte(temp))
	return hex.EncodeToString(sum[:])
}

func formsCacheKey(page, size int, ownerType dto.PermissionType, owner uint64) string {
	uid, _ := auth.UserID()
	temp := fmt.Sprintf("%d%d%s%d%s", page, size, ownerType, owner, uid)
	sum := md5.Sum([]byte(temp))
	return hex.EncodeToString(sum[:])
}

func findFormsFromCache(ctx context.Context, page, size int, ownerType dto.PermissionType, owner uint64) (*dto.GetFormsResponse, error) {
	key := formsCacheKey(page, size, ownerType, owner)
	ans, err := formsCache.Get(ctx, key)
	return &ans, err
}

func formsToDto(v ...*models.Form) (ans []dto.FormConfig) {
	for _, f := range v {
		var bgColor, bgImage, image, description *string
		var deadline *time.Time
		var maxResponses, maxSubmissions *uint

		if f.Deadline.Valid {
			deadline = &f.Deadline.Time
		}

		if f.BackgroundColor.Valid {
			bgColor = &f.BackgroundColor.String
		}

		if f.BackgroundImage.Valid {
			bgImage = &f.BackgroundImage.String
		}

		if f.Image.Valid {
			image = &f.Image.String
		}

		if f.Description.Valid {
			description = &f.Description.String
		}

		if f.MaxResponses.Valid {
			tmp := uint(f.MaxResponses.Int32)
			maxResponses = &tmp
		}

		if f.MaxSubmissions.Valid {
			tmp := uint(f.MaxResponses.Int32)
			maxSubmissions = &tmp
		}

		tmp := dto.FormConfig{
			Id:              f.Id,
			Title:           f.Title,
			CreatedAt:       f.CreatedAt,
			UpdateAt:        f.UpdatedAt,
			MultiResponse:   f.MultiResponse,
			Resubmission:    f.Resubmission,
			Status:          f.Status,
			Description:     description,
			BackgroundColor: bgColor,
			BackgroundImage: bgImage,
			Image:           image,
			Deadline:        deadline,
			MaxResponses:    maxResponses,
			MaxSubmissions:  maxSubmissions,
			Tags:            f.Tags,
			GroupIds:        f.GroupIds,
			QuestionIds:     f.QuestionIds,
		}
		ans = append(ans, tmp)
	}
	return
}

func formQuestionsToDto(f []*models.FormQuestion) []dto.FormQuestion {
	var ans = make([]dto.FormQuestion, len(f))
	for i, q := range f {
		ans[i] = *formQuestionToDto(q)
	}
	return ans
}

func formQuestionGroupsToDto(f []*models.FormQuestionGroup) []dto.FormQuestionGroup {
	var ans = make([]dto.FormQuestionGroup, len(f))
	for i, g := range f {
		ans[i] = *formQuestionGroupToDto(g)
	}
	return ans
}

func formQuestionGroupToDto(f *models.FormQuestionGroup) *dto.FormQuestionGroup {
	var ans = new(dto.FormQuestionGroup)
	ans.Id = f.Id
	ans.Form = f.Form
	if f.Label.Valid {
		ans.Label = &f.Label.String
	}
	if f.Description.Valid {
		ans.Description = &f.Description.String
	}
	if f.Image.Valid {
		ans.Image = &f.Image.String
	}
	return ans
}

func formQuestionToDto(f *models.FormQuestion) *dto.FormQuestion {
	var ans = new(dto.FormQuestion)

	ans.Prompt = f.Prompt
	ans.IsRequired = f.IsRequired
	ans.Type = f.Type
	ans.Id = f.Id
	if f.LayoutVariant.Valid {
		ans.LayoutVariant = f.LayoutVariant.String
	}
	ans.Options = make([]dto.QuestionOption, len(f.Options))

	for i, k := range f.Options {
		ans.Options[i] = dto.QuestionOption{
			Caption:   k.Caption,
			Id:        k.Id,
			Value:     k.Value,
			Image:     k.Image,
			IsDefault: k.IsDefault,
		}
	}

	return ans
}

func updateForm(ctx context.Context, tx *sqldb.Tx, formId uint64, update dto.UpdateFormRequest) error {
	query := `
		UPDATE
			forms
		SET
			title=$1,
			description=$2,
			meta_background=$3,
			meta_bg_img=$4,
			meta_img=$5,
			multi_response=$6,
			response_resubmission=$7,
			deadline=$9,
			max_responses=$10,
			max_submissions=$11
		WHERE
			id = $8;
	`
	res, err := tx.Exec(ctx, query, &update.Title, &update.Description, &update.BackgroundColor, &update.BackgroundImage, &update.Image, &update.MultiResponse, &update.Resubmission, &formId, update.Deadline, update.MaxResponses, update.MaxSubmissions)
	if err != nil {
		return err
	} else if res.RowsAffected() > 0 {
		return bumpFormTimestamp(ctx, tx, formId)
	}

	return nil
}

func createFormQuestion(ctx context.Context, tx *sqldb.Tx, formId uint64, req dto.UpdateFormQuestionRequest) error {
	formExistsQuery := `
		SELECT
			COUNT(id)
		FROM
			forms
		WHERE
			id = $1;
	`

	var formCount int
	if err := tx.QueryRow(ctx, formExistsQuery, formId).Scan(&formCount); err != nil {
		return err
	}

	if formCount == 0 {
		return errs.B().Code(errs.NotFound).Msg("form not found").Err()
	}

	questionInsertQuery := `
		INSERT INTO
			form_questions(form,prompt,is_required,type,layout_variant,form_group)
		VALUES
			($1,$2,$3,$4,$5,$6);
	`

	var group *uint64
	if req.Group > 0 {
		group = &req.Group
	}
	if _, err := tx.Exec(ctx, questionInsertQuery, formId, req.Prompt, req.IsRequired, req.Type, req.LayoutVariant, group); err != nil {
		return err
	}

	return bumpFormTimestamp(ctx, tx, formId)
}

func updateFormQuestion(ctx context.Context, tx *sqldb.Tx, formId uint64, questionId uint64, req dto.UpdateFormQuestionRequest) error {
	updateQuery := `
		UPDATE
			form_questions
		SET
			prompt=$1,
			type=$2,
			layout_variant=$3,
			form_group=$6
		WHERE
			form=$4 AND id=$5
			AND (
				prompt IS DISTINCT FROM $1 OR
				type IS DISTINCT FROM $2 OR
				layout_variant IS DISTINCT FROM $3 OR
				form_group IS DISTINCT FROM $6
			);
	`
	var group *uint64
	if req.Group > 0 {
		group = &req.Group
	}
	res, err := tx.Exec(ctx, updateQuery, req.Prompt, req.Type, req.LayoutVariant, formId, questionId, group)
	if err != nil {
		return err
	}

	if res.RowsAffected() > 0 {
		return bumpFormTimestamp(ctx, tx, formId)
	}

	return nil
}

func updateFormQuestionOptions(ctx context.Context, tx *sqldb.Tx, formId, questionId uint64, req ...dto.FormQuestionOptionUpdate) error {
	for _, v := range req {
		updateQuery := `
			UPDATE
				form_question_options
			SET
				caption=$1,
				value=$2,
				image=$3,
				is_default=$6
			WHERE
				question=$4 AND id=$5
				AND (
					value IS DISTINCT FROM $2 OR
					caption IS DISTINCT FROM $1 OR
					image IS DISTINCT FROM $3 OR
					is_default IS DISTINCT FROM $6
				);
		`
		res, err := tx.Exec(ctx, updateQuery, v.Caption, v.Value, v.Image, questionId, v.Id, v.IsDefault)
		if err != nil {
			return err
		}
		if res.RowsAffected() > 0 {
			return bumpFormTimestamp(ctx, tx, formId)
		}
	}
	return nil
}

func createFormQuestionOptions(ctx context.Context, tx *sqldb.Tx, formId, questionId uint64, req ...dto.NewQuestionOption) error {
	for _, v := range req {
		query := `
			INSERT INTO
				form_question_options(caption,value,image,question,is_default)
			VALUES
				($1,$2,$3,$4,$5)
			RETURNING id;
		`
		var optionId uint64
		if err := tx.QueryRow(ctx, query, v.Caption, v.Value, v.Image, questionId, v.IsDefault).Scan(&optionId); err != nil {
			return err
		} else if v.IsDefault {
			updateQuery := `
				UPDATE
					form_question_options
				SET
					is_default = false
				WHERE
					question=$1 AND id != $2;
			`
			if _, err := tx.Exec(ctx, updateQuery, questionId, optionId); err != nil {
				return err
			}
		}
	}
	return bumpFormTimestamp(ctx, tx, formId)
}

func deleteFormQuestionOptions(ctx context.Context, tx *sqldb.Tx, formId, questionId uint64, req ...uint64) error {
	for _, id := range req {
		query := `
			DELETE FROM
				form_question_options
			WHERE
				id=$1 AND question=$2;
		`
		if _, err := tx.Exec(ctx, query, id, questionId); err != nil {
			return err
		}
	}

	return bumpFormTimestamp(ctx, tx, formId)
}

func deleteFormQuestions(ctx context.Context, tx *sqldb.Tx, formId uint64, req ...uint64) error {
	for _, id := range req {
		query := `
			DELETE FROM
				form_questions
			WHERE
				id=$1 AND form=$2;
		`
		if _, err := tx.Exec(ctx, query, id, formId); err != nil {
			return err
		}
	}
	return bumpFormTimestamp(ctx, tx, formId)
}

func deleteForm(ctx context.Context, tx *sqldb.Tx, formId uint64) error {
	query := `
		DELETE FROM
			forms
		WHERE
			id=$1;
	`
	if _, err := tx.Exec(ctx, query, formId); err != nil {
		return err
	}
	return nil
}

func toggleFormStatus(ctx context.Context, tx *sqldb.Tx, formId uint64) error {
	query := `
		SELECT
			status
		FROM
			forms
		WHERE
			id = $1;
	`

	var currentStatus string
	if err := tx.QueryRow(ctx, query, formId).Scan(&currentStatus); err != nil {
		return err
	}

	var newStatus string
	if currentStatus == "draft" {
		newStatus = "published"
	} else {
		newStatus = "draft"
	}

	query = `
		UPDATE
			forms
		SET
			status=$1
		WHERE
			id=$2;
	`

	if _, err := tx.Exec(ctx, query, newStatus, formId); err != nil {
		return err
	}

	return bumpFormTimestamp(ctx, tx, formId)
}

func createQuestionsGroup(ctx context.Context, tx *sqldb.Tx, formId uint64, req dto.UpdateFormQuestionGroupRequest) error {
	query := `
		INSERT INTO
			form_question_groups(form,label,description,image)
		VALUES
			($1,$2,$3,$4);
	`

	var label, description, image *string
	if req.Label != nil && len(*req.Label) > 0 {
		label = req.Label
	}

	if req.Description != nil && len(*req.Description) > 0 {
		description = req.Description
	}

	if req.Image != nil && len(*req.Image) > 0 {
		image = req.Image
	}

	if _, err := tx.Exec(ctx, query, formId, label, description, image); err != nil {
		return err
	}

	return bumpFormTimestamp(ctx, tx, formId)
}

func bumpFormTimestamp(ctx context.Context, tx *sqldb.Tx, formId uint64) error {
	query := `
		UPDATE
			forms
		SET
			updated_at=DEFAULT
		WHERE
			id=$1;
	`
	if _, err := tx.Exec(ctx, query, formId); err != nil {
		return err
	}
	return nil
}

func updateFormQuestionGroup(ctx context.Context, tx *sqldb.Tx, form, group uint64, req dto.UpdateFormQuestionGroupRequest) error {

	query := `
		UPDATE
			form_question_groups
		SET
			label=$1,
			description=$2,
			image=$3
		WHERE
			id=$4 AND form=$5
			AND (
				label IS DISTINCT FROM $1 OR
				description IS DISTINCT FROM $2 OR
				image IS DISTINCT FROM $3
			);
	`

	res, err := tx.Exec(ctx, query, req.Label, req.Description, req.Image, group, form)
	if err != nil {
		return err
	}
	if res.RowsAffected() > 0 {
		return bumpFormTimestamp(ctx, tx, form)
	}
	return nil
}

func deleteFormQuestionGroups(ctx context.Context, tx *sqldb.Tx, form uint64, ids ...uint64) error {
	groupIds := ""
	for i, id := range ids {
		groupIds += fmt.Sprintf("%d", id)
		if i < len(ids)-1 {
			groupIds += ","
		}
	}
	if len(groupIds) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		DELETE FROM
			form_question_groups
		WHERE
			form=$1 AND id IN (%s)
	`, groupIds)

	res, err := tx.Exec(ctx, query, form)
	if err != nil {
		return err
	}

	if res.RowsAffected() > 0 {
		return bumpFormTimestamp(ctx, tx, form)
	}

	return nil
}

func findUserResponses(ctx context.Context, user, form uint64) ([]*models.FormResponse, error) {
	query := `
		SELECT
			fr.id,
			fr.responder,
			fr.submitted_at,
			fr.created_at,
			fr.updated_at,
			COALESCE(json_agg(json_build_object(
				'id', ra.id,
				'question', ra.question,
				'value', ra.value,
				'response', ra.response
			)) FILTER (WHERE ra.response IS NOT NULL), '[]') as answers,
			fr.form
		FROM
			form_responses fr
		LEFT JOIN
			response_answers ra
				ON ra.response=fr.id
		WHERE
			fr.responder=$1 AND fr.form=$2
		GROUP BY
			fr.id;
	`

	rows, err := formsDb.Query(ctx, query, user, form)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses = make([]*models.FormResponse, 0)
	for rows.Next() {
		var response = new(models.FormResponse)
		var answersJson string
		if err := rows.Scan(&response.Id, &response.Responder, &response.SubmittedAt, &response.CreatedAt, &response.UpdatedAt, &answersJson, &response.Form); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(answersJson), &response.Answers); err != nil {
			return nil, err
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func responsesToDto(r ...*models.FormResponse) []dto.UserFormResponse {
	var ans = make([]dto.UserFormResponse, len(r))

	for i, v := range r {
		var res dto.UserFormResponse
		res.Id = v.Id
		res.CreatedAt = v.CreatedAt
		res.UpdatedAt = v.UpdatedAt
		res.Responsder = v.Responder
		if v.SubmittedAt.Valid {
			res.SubmittedAt = &v.SubmittedAt.Time
		}

		res.Answers = make([]dto.FormAnswer, len(v.Answers))
		for j, w := range v.Answers {
			var a dto.FormAnswer
			a.Id = w.Id
			a.Question = w.Question
			a.Response = w.Response
			a.CreatedAt = w.CreatedAt
			a.UpdatedAt = w.UpdatedAt
			a.Value = w.Value

			res.Answers[j] = a
		}

		ans[i] = res
	}

	return ans
}

func createUserFormResponse(ctx context.Context, tx *sqldb.Tx, form, user uint64) error {
	query := `
		INSERT INTO
			form_responses(responder,form)
		VALUES
			($1,$2);
	`

	if _, err := tx.Exec(ctx, query, user, form); err != nil {
		return err
	}
	return nil
}

func countUserResponses(ctx context.Context, form, user uint64) (uint, uint, error) {
	query := `
		SELECT
			COUNT(id),
			(
				SELECT 
					COUNT(id) 
				FROM
					form_responses
				WHERE
					responder=$1 AND form=$2 AND submitted_at IS NOT NULL
			)
		FROM
			form_responses
		WHERE
			responder=$1 AND form=$2;
	`

	var total, submitted uint
	if err := formsDb.QueryRow(ctx, query, user, form).Scan(&total, &submitted); err != nil {
		return 0, 0, err
	}
	return total, submitted, nil
}

func findUserResponseById(ctx context.Context, form, user, response uint64) (*models.FormResponse, error) {
	query := `
		SELECT
			fr.id,
			fr.responder,
			fr.submitted_at,
			fr.created_at,
			fr.updated_at,
			COALESCE(json_agg(json_build_object(
				'id', ra.id,
				'question', ra.question,
				'value', ra.value,
				'response', ra.response
			)) FILTER (WHERE ra.response IS NOT NULL), '[]') as answers,
			fr.form
		FROM
			form_responses fr
		LEFT JOIN
			response_answers ra
				ON ra.response=fr.id
		WHERE
			fr.responder=$1 AND fr.form=$2 AND fr.id=$3
		GROUP BY
			fr.id;
	`

	var r = new(models.FormResponse)
	var answersJson string
	if err := formsDb.QueryRow(ctx, query, user, form, response).Scan(&r.Id, &r.Responder, &r.SubmittedAt, &r.CreatedAt, &r.UpdatedAt, &answersJson, &r.Form); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(answersJson), &r.Answers); err != nil {
		return nil, err
	}

	return r, nil
}

func bumpFormResponseTimestamp(ctx context.Context, tx *sqldb.Tx, id uint64) error {
	query := `
		UPDATE
			form_responses
		SET
			updated_at=DEFAULT
		WHERE
			id=$1;
	`
	if _, err := tx.Exec(ctx, query, id); err != nil {
		return err
	}
	return nil
}

type questionConstraints struct {
	IsRequired bool
}

func updateUserResponseAnswers(ctx context.Context, tx *sqldb.Tx, form, response uint64, req ...dto.FormAnswerUpdate) error {
	args := make([]any, len(req)+1)
	placeholders := make([]string, len(req))
	args[0] = form

	for i, v := range req {
		args[i+1] = v.Question
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			is_required
		FROM
			form_questions
		WHERE
			form=$1 AND id IN (%s);
	`, strings.Join(placeholders, ","))

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var constraints = make(map[uint64]questionConstraints)
	for rows.Next() {
		var constraint questionConstraints
		var id uint64
		if err := rows.Scan(&id, &constraint.IsRequired); err != nil {
			return err
		}
		constraints[id] = constraint
	}

	answersQuery := `
		SELECT
			question,id
		FROM
			response_answers
		WHERE
			response=$1;
	`

	var qs = make(map[uint64]uint64)
	rows, err = tx.Query(ctx, answersQuery, response)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var q, id uint64
		if err := rows.Scan(&q, &id); err != nil {
			return err
		}
		qs[q] = id
	}

	for _, v := range req {
		constraint, ok := constraints[v.Question]
		if !ok {
			return &errs.Error{
				Code:    errs.Aborted,
				Message: fmt.Sprintf("No question with ID = %d", v.Question),
			}
		}

		if constraint.IsRequired && v.Value == nil || v.Value != nil && len(*v.Value) == 0 {
			return &errs.Error{
				Code:    errs.Aborted,
				Message: fmt.Sprintf("Question with ID=%d is required", v.Question),
			}
		}

		answerId, ok := qs[v.Question]
		if ok {
			updateQuery := `
				UPDATE
					response_answers
				SET
					value=$1,
					updated_at=DEFAULT
				WHERE
					response=$2 AND question=$3 AND id=$4;
			`
			if _, err := tx.Exec(ctx, updateQuery, v.Value, response, v.Question, answerId); err != nil {
				return err
			}
		} else {
			insertQuery := `
				INSERT INTO
					response_answers(value,question,response)
				VALUES
					($1,$2,$3);				
			`
			if _, err := tx.Exec(ctx, insertQuery, v.Value, v.Question, response); err != nil {
				return err
			}
		}
	}

	return bumpFormResponseTimestamp(ctx, tx, response)
}

func verifyUserOwnedResponse(ctx context.Context, user, response, form uint64) bool {
	query := `
		SELECT
			COUNT(id)
		FROM
			form_responses
		WHERE
			id=$1 AND responder=$2 AND form=$3
	`

	var cnt int
	if err := formsDb.QueryRow(ctx, query, response, user, form).Scan(&cnt); err != nil {
		return false
	}
	return cnt > 0
}

func deleteResponseAnswers(ctx context.Context, tx *sqldb.Tx, form, response uint64, req ...uint64) error {
	args := make([]any, len(req)+2)
	var placeholders = make([]string, len(args))
	args[0], args[1], placeholders[0], placeholders[1] = response, form, "$1", "$2"

	for i, v := range req {
		args[i+2], placeholders[i+2] = v, fmt.Sprintf("$%d", i+3)
	}

	query := fmt.Sprintf(`
		DELETE FROM
			response_answers
		WHERE
			response=$1 AND form=$2 AND question IN (%s);
	`, strings.Join(placeholders, ","))

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return err
	}
	return bumpFormResponseTimestamp(ctx, tx, response)
}

func submitUserResponse(ctx context.Context, tx *sqldb.Tx, form, user, response uint64) error {

	invalidAnswersQuery := `
		SELECT
			COUNT(ra.id)
		FROM
			response_answers ra
		JOIN
			form_questions fq
				ON ra.question=fq.id
		WHERE
			ra.response=$1 AND 
			fq.form=$2 AND 
			fq.is_required=true AND 
			ra.value IS NULL;
	`

	var cnt int
	if err := tx.QueryRow(ctx, invalidAnswersQuery, response, form).Scan(&cnt); err != nil {
		return err
	} else if cnt > 0 {
		return &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "Cannot submit. There are invalid answers in this submission",
		}
	}

	query := `
		UPDATE
			form_responses
		SET
			submitted_at=NOW()
		WHERE
			responder=$1 AND submitted_at IS NULL AND form=$2 AND id=$3;
	`

	res, err := tx.Exec(ctx, query, user, form, response)
	if err != nil {
		return err
	} else if res.RowsAffected() > 0 {
		return bumpFormResponseTimestamp(ctx, tx, response)
	}
	return nil
}
