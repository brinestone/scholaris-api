package util

import (
	"encore.dev/beta/errs"
)

const (
	MsgDbAccessError    = "db access error"
	MsgCacheAccessError = "cache access error"
	MsgWebhookError     = "webhook error"
	MsgForbidden        = "forbidden action"
	MsgCallError        = "error while calling API"
	MsgUploadError      = "error while uploading file"
)

var ErrConflict = errs.Error{
	Code:    errs.AlreadyExists,
	Message: "Duplicate resource",
}

var ErrUnknown = errs.Error{
	Code:    errs.Internal,
	Message: "Internal server error",
}

var ErrNotFound = errs.Error{
	Code:    errs.NotFound,
	Message: "Resource not found",
}

var ErrForbidden = errs.Error{
	Code:    errs.PermissionDenied,
	Message: "Permission not allowed",
}

var ErrCaptchaError = errs.Error{
	Code:    errs.FailedPrecondition,
	Message: "reCaptcha Verification failed",
}

var ErrUnauthorized = errs.Error{
	Code:    errs.Unauthenticated,
	Message: "Unauthorized",
}
