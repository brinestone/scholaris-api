package forms

import (
	encoreAuth "encore.dev/beta/auth"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

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
