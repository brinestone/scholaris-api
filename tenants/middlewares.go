package tenants

import (
	eAuth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

//encore:middleware target=tag:can_view_tenant
func CanViewTenant(req middleware.Request, next middleware.Next) (res middleware.Response) {
	res = next(req)

	uid, _ := eAuth.UserID()
	if len(uid) == 0 {
		res = middleware.Response{
			Err: &util.ErrUnauthorized,
		}
		return
	}

	perm, err := permissions.CheckPermissionInternal(req.Context(), dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: dto.PNCanView,
		Target:   dto.IdentifierString(dto.PTTenant, req.Data().PathParams.Get("id")),
	})

	if err != nil {
		err = errs.Wrap(err, util.MsgCallError, "middleware", "CanViewTenant", "msg", err.Error())
		rlog.Error("middleware error", err.Error())
		res = middleware.Response{
			Err: &util.ErrUnknown,
		}
		return
	}

	if !perm.Allowed {
		res = middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	return
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
