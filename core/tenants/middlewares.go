package tenants

import (
	"encore.dev/middleware"
	"encore.dev/rlog"
)

//encore:middleware target=tag:perm_can_delete_tenant
func AllowedToDeleteInstitutionMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
	rlog.Debug("user is allowed to delete")
	return next(req)
}
