package institutions

import (
	"fmt"

	"encore.dev/beta/auth"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/util"
)

// Validates a user's permission to create an institution
//
//encore:middleware target=tag:perm_can_create_institution
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
		Subject:  fmt.Sprintf("user:%v", userId),
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
