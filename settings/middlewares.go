package settings

import (
	"context"

	eAuth "encore.dev/beta/auth"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Verifies whethe the user can set the value of a setting.
//
//encore:middleware target=tag:can_set_setting
func UserCanSetSettingValue(req middleware.Request, next middleware.Next) middleware.Response {
	var ownerInfo = req.Data().Payload.(models.OwnerInfo)
	if err := doPermissionCheck(req.Context(), ownerInfo, dto.PermCanSetSettingValue); err != nil {
		return middleware.Response{
			Err: err,
		}
	}

	return next(req)
}

// Verifies whether the user can update an owner's settings.
//
//encore:middleware target=tag:can_update_settings
func UserCanUpdateSettings(req middleware.Request, next middleware.Next) middleware.Response {
	var ownerInfo = req.Data().Payload.(models.OwnerInfo)
	if err := doPermissionCheck(req.Context(), ownerInfo, dto.PermCanEditSettings); err != nil {
		return middleware.Response{
			Err: err,
		}
	}

	return next(req)
}

// Verifies whether the user can view an owner's settings.
//
//encore:middleware target=tag:can_view_settings
func UserCanViewSettings(req middleware.Request, next middleware.Next) middleware.Response {
	var ownerInfo = req.Data().Payload.(models.OwnerInfo)
	if err := doPermissionCheck(req.Context(), ownerInfo, dto.PermCanViewSettings); err != nil {
		return middleware.Response{
			Err: err,
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

func doPermissionCheck(ctx context.Context, oi models.OwnerInfo, perm dto.PermissionName) error {
	uid, authed := eAuth.UserID()
	if !authed {
		return &util.ErrForbidden
	}

	parsed, _ := dto.ParsePermissionType(oi.GetOwnerType())
	perms, err := permissions.CheckPermissionInternal(ctx, dto.InternalRelationCheckRequest{
		Relation: perm,
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Target:   dto.IdentifierString(parsed, oi.GetOwner()),
	})

	if err != nil {
		rlog.Warn(util.MsgCallError, "msg", err.Error())
	}

	if perms != nil && !perms.Allowed {
		return &util.ErrForbidden
	} else if perms == nil {
		return &util.ErrUnknown
	}

	return nil
}
