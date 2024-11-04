// CRUD endpoints for forms
package forms

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Updates a form question's options
//
//encore:api auth method=PATCH path=/forms/:form/questions/:question/options
func UpdateFormQuestionOptions(ctx context.Context, form uint64, question uint64, req dto.UpdateFormQuestionOptionsRequest) (*dto.GetFormQuestionsResponse, error) {

	return nil, nil
}

// Updates a form question
//
//encore:api auth method=PATCH path=/forms/:form/questions/:question
func UpdateQuestion(ctx context.Context, form uint64, question uint64) error {
	return nil
}

// Add a question to a form
//
//encore:api auth method=POST path=/forms/:formId/question
func CreateQuestion(ctx context.Context, formId uint64, req dto.NewFormQuestionRequest) (*dto.GetFormQuestionsResponse, error) {
	uid, _ := auth.UserID()
	perm, err := permissions.CheckPermission(ctx, dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermEditor,
		Target:   dto.IdentifierString(dto.PTForm, formId),
	})
	if err != nil {
		rlog.Error(err.Error())
	}
	if perm == nil || !perm.Allowed {
		return nil, &util.ErrForbidden
	}

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := createFormQuestion(ctx, tx, formId, req); err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		}
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	questions, err := findFormQuestionsFromDb(ctx, formId)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	ans := dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions)}
	if err := questionsCache.Set(ctx, formId, ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return &ans, nil
}

// Update a form
//
//encore:api auth method=PUT path=/forms/:id
func UpdateForm(ctx context.Context, id uint64, req dto.UpdateFormRequest) (*dto.FormConfig, error) {
	uid, _ := auth.UserID()
	perm, err := permissions.CheckPermission(ctx, dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermEditor,
		Target:   dto.IdentifierString(dto.PTForm, id),
	})
	if err != nil {
		rlog.Error(err.Error())
	}
	if perm == nil || !perm.Allowed {
		return nil, &util.ErrForbidden
	}

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

	form, err := findFormFromDbTx(ctx, tx, id)
	if err != nil {
		rlog.Error(err.Error())
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	ans := formToDto(form)
	if err := formCache.Set(ctx, id, *ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return ans, nil
}

// Gets an owner's form data
//
//encore:api public method=GET path=/forms/:id/questions
func FindFormQuestions(ctx context.Context, id uint64) (*dto.GetFormQuestionsResponse, error) {
	response, err := questionsCache.Get(ctx, id)
	if errors.Is(err, cache.Miss) {
		questions, err := findFormQuestionsFromDb(ctx, id)
		if err != nil {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}

		response = dto.GetFormQuestionsResponse{Questions: formQuestionsToDto(questions)}

		if len(response.Questions) > 0 {
			if err := questionsCache.Set(ctx, id, response); err != nil {
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
func FindForms(ctx context.Context, params dto.GetFormsInput) (*dto.PaginatedResponse[dto.FormConfig], error) {
	ownerType, _ := dto.PermissionTypeFromString(params.OwnerType)
	res, err := findFormsFromCache(ctx, int(params.Page), int(params.Size), ownerType, params.Owner)
	if errors.Is(err, cache.Miss) {
		formsFromDb, count, err := findFormsFromDb(ctx, int(params.Page), int(params.Size), params.Owner)
		if err != nil {
			return nil, err
		}

		var forms []*dto.FormConfig = make([]*dto.FormConfig, len(formsFromDb))
		for i, v := range formsFromDb {
			forms[i] = formToDto(v)
		}

		meta := dto.PaginatedResponseMeta{
			Total: uint(count),
		}

		response := &dto.PaginatedResponse[dto.FormConfig]{
			Data: forms,
			Meta: meta,
		}

		if len(formsFromDb) > 0 {
			key := formsCacheKey(int(params.Page), int(params.Size), ownerType, params.Owner)
			if err := formsCache.Set(ctx, key, *response); err != nil {
				rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
			}
		}
		return response, nil
	} else if err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	return res, nil
}

// Creates a new form
//
//encore:api auth method=POST path=/forms tag:needs_captcha_ver
func NewForm(ctx context.Context, req dto.NewFormInput) (*dto.FormConfig, error) {
	uid, _ := auth.UserID()
	pt, _ := dto.PermissionTypeFromString(req.OwnerType)
	permission, err := permissions.CheckPermission(ctx, dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermCanCreateForms,
		Target:   dto.IdentifierString(pt, req.Owner),
	})
	if err != nil {
		if errs.Convert(err) != nil {
			rlog.Warn(util.MsgCallError, "msg", err.Error())
		}
		return nil, &util.ErrForbidden
	} else if !permission.Allowed {
		return nil, &util.ErrForbidden
	}

	tx, err := formsDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	form, err := createForm(ctx, tx, req.Owner, req)
	if err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		} else {
			rlog.Error(err.Error())
			return nil, &util.ErrUnknown
		}
	}

	if err := permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Subject:  dto.IdentifierString(dto.PermissionType(req.OwnerType), req.Owner),
				Relation: models.PermOwner,
				Target:   dto.IdentifierString(dto.PTForm, form.Id),
			},
		},
	}); err != nil {
		rlog.Error(err.Error())
		tx.Rollback()
		return nil, &util.ErrUnknown
	}
	tx.Commit()

	ans := formToDto(form)
	if err := formCache.Set(ctx, form.Id, *ans); err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
	}

	NewForms.Publish(ctx, FormCreated{
		Id:        ans.Id,
		Timestamp: ans.CreatedAt,
	})
	return ans, nil
}

func createForm(ctx context.Context, tx *sqldb.Tx, owner uint64, input dto.NewFormInput) (*models.Form, error) {
	query := `
		INSERT INTO
			forms(title,description,meta_background,meta_bg_img,meta_img,owner,multi_response,response_resubmission)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id;
	`

	var description, bg, bgImg, img *string
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

	var id uint64
	if err := tx.QueryRow(ctx, query, &input.Title, description, bg, bgImg, img, &owner, &input.MultiResponse, &input.Resubmission).Scan(&id); err != nil {
		return nil, err
	}

	return findFormFromDbTx(ctx, tx, id)
}

func findFormQuestionsFromDb(ctx context.Context, id uint64) ([]*models.FormQuestion, error) {
	query := `
		SELECT
			fq.id,
			fq.prompt,
			fq.is_required,
			fq.type,
			fq.layout_variant,
			COALESCE(json_agg(json_build_obj(
				'id', fqo.id,
				'caption', fqo.caption,
				'value', fqo.caption,
				'image', fqo.image,
			)) FILTER (WHERE fqo.question IS NOT NULL), '[]') AS options
		FROM
			form_questions fq
		LEFT JOIN
			form_question_options fqo
				ON fqo.question = fq.id
		WHERE
			fq.form = $1
		GROUP BY
			fq.id;
	`

	rows, err := formsDb.Query(ctx, query, id)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ans []*models.FormQuestion = make([]*models.FormQuestion, 0)
	for rows.Next() {
		var question = new(models.FormQuestion)
		var optionsJson string
		if err := rows.Scan(&question.Id, &question.Prompt, &question.IsRequired, &question.Type, &question.LayoutVariant, &optionsJson); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(optionsJson), &question.Options); err != nil {
			return nil, err
		}

		ans = append(ans, question)
	}

	return ans, nil
}

func findFormFromDbTx(ctx context.Context, tx *sqldb.Tx, id uint64) (*models.Form, error) {
	query := `
		SELECT
			id,
			title,
			description,
			meta_background,
			meta_bg_img,
			meta_img,
			created_at,
			updated_at,
			owner,
			multi_response,
			response_resubmission,
			status
		FROM
			forms
		WHERE
			id = $1
		;
	`

	var form *models.Form = new(models.Form)
	if err := tx.QueryRow(ctx, query, id).Scan(&form.Id, &form.Title, &form.Description, &form.BackgroundColor, &form.BackgroundImage, &form.Image, &form.CreatedAt, &form.UpdatedAt, &form.Owner, &form.MultiResponse, &form.Resubmission, &form.Status); err != nil {
		return nil, err
	}
	return form, nil
}

func findFormsFromDb(ctx context.Context, page, size int, owner uint64) ([]*models.Form, uint64, error) {
	query := `
		SELECT
			id,
			title,
			description,
			meta_background,
			meta_bg_img,
			meta_img,
			created_at,
			updated_at,
			owner,
			multi_response,
			response_resubmission,
			status
		FROM
			forms
		WHERE
			owner = $1
		OFFSET $2
		LIMIT $3;
	`

	rows, err := formsDb.Query(ctx, query, owner, page*size, size)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ans = make([]*models.Form, 0)
	for rows.Next() {
		var form = new(models.Form)
		if err := rows.Scan(&form.Id, &form.Title, &form.Description, &form.BackgroundColor, &form.BackgroundImage, &form.Image, &form.CreatedAt, &form.UpdatedAt, &form.Owner, &form.MultiResponse, &form.Resubmission, &form.Status); err != nil {
			return nil, 0, err
		}
		ans = append(ans, form)
	}

	countQuery := `
		SELECT
			COUNT(*)
		FROM
			forms
		WHERE
			owner = $1;
	`
	var count uint64
	if err := formsDb.QueryRow(ctx, countQuery, owner).Scan(&count); err != nil {
		return nil, 0, err
	}
	return ans, count, nil
}

func formsCacheKey(page, size int, ownerType dto.PermissionType, owner uint64) string {
	temp := fmt.Sprintf("%d%d%s%d", page, size, ownerType, owner)
	sum := md5.Sum([]byte(temp))
	return hex.EncodeToString(sum[:])
}

func findFormsFromCache(ctx context.Context, page, size int, ownerType dto.PermissionType, owner uint64) (*dto.PaginatedResponse[dto.FormConfig], error) {
	key := formsCacheKey(page, size, ownerType, owner)
	ans, err := formsCache.Get(ctx, key)
	return &ans, err
}

func formToDto(f *models.Form) *dto.FormConfig {
	var bgColor, bgImage, image, description *string

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

	return &dto.FormConfig{
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
	}
}

func formQuestionsToDto(f []*models.FormQuestion) []dto.FormQuestion {
	var ans = make([]dto.FormQuestion, len(f))
	for i, q := range f {
		ans[i] = *formQuestionToDto(q)
	}
	return ans
}

func formQuestionToDto(f *models.FormQuestion) *dto.FormQuestion {
	var ans = new(dto.FormQuestion)

	ans.Prompt = f.Prompt
	// ans.ResponseType = f.ResponseType
	// ans.Form = f.Form
	ans.IsRequired = f.IsRequired
	ans.Id = f.Id
	if f.LayoutVariant.Valid {
		ans.LayoutVariant = f.LayoutVariant.String
	}
	ans.Options = make([]dto.QuestionOption, len(f.Options))

	for i, k := range f.Options {
		ans.Options[i] = dto.QuestionOption{
			Caption: k.Caption,
			Id:      k.Id,
		}
		if k.Value.Valid {
			ans.Options[i].Value = &k.Value.String
		}
		if k.Image.Valid {
			ans.Options[i].Image = &k.Image.String
		}
	}

	return ans
}

func updateForm(ctx context.Context, tx *sqldb.Tx, formId uint64, update dto.UpdateFormRequest) error {
	query := `
		UPDATE
			forms
		SET
			updated_at=DEFAULT,
			title=$1,
			description=$2,
			meta_background=$3,
			meta_bg_img=$4,
			meta_img=$5,
			multi_response=$6,
			response_resubmission=$7
		WHERE
			id = $8;
	`

	if _, err := tx.Exec(ctx, query, &update.Title, &update.Description, &update.BackgroundColor, &update.BackgroundImage, &update.Image, &update.MultiResponse, &update.Resubmission, &formId); err != nil {
		return err
	}
	return nil
}

func createFormQuestion(ctx context.Context, tx *sqldb.Tx, formId uint64, req dto.NewFormQuestionRequest) error {
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
			form_questions(form,prompt,is_required,type,layout_variant)
		VALUES
			($1,$2,$3,$4,$5);
	`

	if _, err := tx.Exec(ctx, questionInsertQuery, formId, req.Prompt, req.IsRequired, req.Type, req.LayoutVariant); err != nil {
		return err
	}

	return nil
}
