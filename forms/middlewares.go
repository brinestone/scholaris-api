package forms

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	encoreAuth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

//encore:middleware target=tag:user_owns_response
func UserOwnsResponse(req middleware.Request, next middleware.Next) middleware.Response {
	ctx := req.Context()
	formId := req.Data().PathParams.Get("form")
	form, _ := strconv.ParseUint(formId, 10, 64)
	responseId := req.Data().PathParams.Get("response")
	response, _ := strconv.ParseUint(responseId, 10, 64)
	sub, _ := encoreAuth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)

	if !verifyUserOwnedResponse(ctx, uid, response, form) {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}
	return next(req)
}

//encore:middleware target=tag:user_can_submit_response
func CanUserSubmitResponse(req middleware.Request, next middleware.Next) middleware.Response {
	ctx := req.Context()
	formId := req.Data().PathParams.Get("form")
	form, _ := strconv.ParseUint(formId, 10, 64)

	sub, _ := encoreAuth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)
	_, submitted, err := countUserResponses(ctx, form, uid)
	if err != nil && !errors.Is(err, sqldb.ErrNoRows) {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}

	formConfig, err := findFormFromDb(ctx, form)
	if errors.Is(err, sqldb.ErrNoRows) {
		return middleware.Response{
			Err: &util.ErrNotFound,
		}
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}
	if formConfig.Resubmission && formConfig.MaxSubmissions.Valid && formConfig.MaxSubmissions.Int32 <= int32(submitted) {
		return middleware.Response{
			Err: &errs.Error{
				Code:    errs.ResourceExhausted,
				Message: "Maxiumum number of submissions reached",
			},
		}
	}
	return next(req)
}

//encore:middleware target=tag:user_can_respond_to_form
func CanUserRespondToForm(req middleware.Request, next middleware.Next) middleware.Response {
	ctx := req.Context()
	formId := req.Data().PathParams.Get("form")
	form, _ := strconv.ParseUint(formId, 10, 64)

	sub, _ := encoreAuth.UserID()
	uid, _ := strconv.ParseUint(string(sub), 10, 64)
	total, _, err := countUserResponses(ctx, form, uid)
	if err != nil && !errors.Is(err, sqldb.ErrNoRows) {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}

	formConfig, err := findFormFromDb(ctx, form)
	if errors.Is(err, sqldb.ErrNoRows) {
		return middleware.Response{
			Err: &util.ErrNotFound,
		}
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}

	now := time.Now()
	if formConfig.Status == dto.FSDraft {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	} else if formConfig.Deadline.Valid && formConfig.Deadline.Time.Before(now) {
		return middleware.Response{
			Err: &errs.Error{
				Code:    errs.DeadlineExceeded,
				Message: fmt.Sprintf("Deadline exceeded - %s", formConfig.Deadline.Time.String()),
			},
		}
	} else if formConfig.MultiResponse && formConfig.MaxResponses.Valid && formConfig.MaxResponses.Int32 <= int32(total) {
		return middleware.Response{
			Err: &errs.Error{
				Code:    errs.ResourceExhausted,
				Message: "Maximum number of responses reached",
			},
		}
	}
	return next(req)
}

//encore:middleware target=tag:user_is_form_editor
func VerifyFormEditor(req middleware.Request, next middleware.Next) middleware.Response {
	uid, _ := encoreAuth.UserID()
	form := req.Data().PathParams.Get("form")
	canEdit, err := permissions.CheckPermission(req.Context(), dto.RelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: models.PermEditor,
		Target:   dto.IdentifierString(dto.PTForm, form),
	})
	if err != nil {
		rlog.Error(err.Error())
	}
	if canEdit == nil || !canEdit.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	return next(req)
}

//encore:middleware target=tag:needs_captcha_ver
func VerifyCaptcha(req middleware.Request, next middleware.Next) middleware.Response {
	p, ok := req.Data().Payload.(models.CaptchaVerifiable)
	if !ok {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	if err := auth.VerifyCaptchaToken(req.Context(), auth.VerifyCaptchaRequest{
		Token: p.GetCaptchaToken(),
	}); err != nil {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	return next(req)
}
