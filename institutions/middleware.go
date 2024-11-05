package institutions

import (
	"fmt"

	"encore.dev/beta/auth"
	"encore.dev/middleware"
	"encore.dev/rlog"
	appAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Verifies the captcha token in a request
//
//encore:middleware target=tag:needs_captcha_ver
func VerifyCaptchaTokenMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
	data, ok := req.Data().Payload.(models.CaptchaVerifiable)
	if !ok {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	if err := appAuth.VerifyCaptchaToken(req.Context(), appAuth.VerifyCaptchaRequest{
		Token: data.GetCaptchaToken(),
	}); err != nil {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	return next(req)
}

// Validates a user's permission to create an institution
//
//encore:middleware target=tag:perm_can_create
func AllowedToCreateInstitutionMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
	userId, signedIn := auth.UserID()
	if !signedIn {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	data, ok := req.Data().Payload.(dto.NewInstitutionRequest)
	if !ok {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	ans, err := permissions.CheckPermission(req.Context(), dto.RelationCheckRequest{
		Actor:    fmt.Sprintf("user:%v", userId),
		Relation: "can_create_institution",
		Target:   fmt.Sprintf("tenant:%d", data.TenantId),
	})

	if err != nil {
		rlog.Error(err.Error())
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}

	if !ans.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	return next(req)
}
