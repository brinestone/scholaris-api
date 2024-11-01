// CRUD endpoints for forms
package forms

import (
	"context"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Creates a new form
//
//encore:api auth method=POST path=/forms tag:needs_captcha_ver
func NewForm(ctx context.Context, req dto.NewFormInput) (*dto.FormConfig, error) {
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
	return ans, nil
}

func createForm(ctx context.Context, tx *sqldb.Tx, owner uint64, input dto.NewFormInput) (*models.Form, error) {
	query := `
		INSERT INTO
			forms(title,description,meta_background,meta_bg_img,meta_img,owner,multi_response,response_resubmission)
		VALUES
			($1,$2,$3,$4,$5,$6,$7$8)
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
