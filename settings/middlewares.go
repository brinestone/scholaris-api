package settings

import (
	"strconv"

	eAuth "encore.dev/beta/auth"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Verifies whether the user can update an owner's settings.
//
//encore:middleware target=tag:can_update_settings
func UserCanUpdateSettings(req middleware.Request, next middleware.Next) middleware.Response {
	uid, authed := eAuth.UserID()
	if !authed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	ownerInfo, _ := req.Data().Payload.(models.OwnerInfo)
	parsed, _ := dto.PermissionTypeFromString(ownerInfo.GetOwnerType())
	perms, err := permissions.CheckPermission(req.Context(), dto.RelationCheckRequest{
		Relation: models.PermCanEdit,
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Target:   dto.IdentifierString(parsed, ownerInfo.GetOwner()),
	})

	if err != nil {
		rlog.Warn(util.MsgCallError, "msg", err.Error())
	}

	if perms != nil && !perms.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	} else if perms == nil {
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}

	return next(req)
}

// Verifies whether the user can view an owner's settings.
//
//encore:middleware target=tag:can_view_settings
func UserCanViewSettings(req middleware.Request, next middleware.Next) middleware.Response {
	uid, authed := eAuth.UserID()
	if !authed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	var ownerId uint64
	var ownerType string
	ownerId, _ = strconv.ParseUint(req.Data().Headers.Get("x-owner"), 10, 64)
	ownerType = req.Data().Headers.Get("x-owner-type")
	parsed, _ := dto.PermissionTypeFromString(ownerType)
	perms, err := permissions.CheckPermission(req.Context(), dto.RelationCheckRequest{
		Relation: models.PermCanView,
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Target:   dto.IdentifierString(parsed, ownerId),
	})

	if err != nil {
		rlog.Warn(util.MsgCallError, "msg", err.Error())
	}

	if perms != nil && !perms.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	} else if perms == nil {
		return middleware.Response{
			Err: &util.ErrUnknown,
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
