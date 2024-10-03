package util

import "encore.dev/beta/errs"

var ErrUnknown = errs.Error{
	Code:    errs.Internal,
	Message: "An unknown error occured",
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
