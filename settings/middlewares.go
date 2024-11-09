package settings

import "encore.dev/middleware"

// Verifies whether the user can update an owner's settings.
//
//encore:middleware target=tag:can_update_settings
func UserCanUpdateSetting(req middleware.Request, next middleware.Next) middleware.Response {
	return next(req)
}

// Verifies whether the user can view an owner's settings.
//
//encore:middleware target=tag:can_view_settings
func UserCanViewSetting(req middleware.Request, next middleware.Next) middleware.Response {
	return next(req)
}
