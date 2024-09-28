package util

import "encore.dev/beta/errs"

var ErrUnknown = errs.Error{
	Code:    errs.Internal,
	Message: "An unknown error occured",
}
