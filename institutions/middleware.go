package institutions

import (
	"errors"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/middleware"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	appAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Validates a user's permission to create an enrollment
//
//encore:middleware target=tag:can_enroll
func AllowedToEnroll(request middleware.Request, next middleware.Next) (ans middleware.Response) {
	ans = next(request)

	uid, _ := auth.UserID()
	ownerInfo, ok := request.Data().Payload.(models.OwnerInfo)
	if !ok {
		ans = middleware.Response{
			Err: &util.ErrForbidden,
		}
		return
	}

	lookup, err := findInstitutionByGenericIdentifier(request.Context(), fmt.Sprintf("%d", ownerInfo.GetOwner()))
	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err.Error())
		ans = middleware.Response{
			Err: err,
		}
		return
	}

	levelInfo, ok := request.Data().Payload.(models.HasLevelIdentifier)
	if !ok {
		ans = middleware.Response{
			Err: &util.ErrForbidden,
		}
		return
	}

	formInfo, err := findLevelEnrollmentForm(request.Context(), levelInfo.GetLevelRef(), ownerInfo.GetOwner())
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			rlog.Warn("enrollment attempted", "reason", "no form configured", "level", levelInfo.GetLevelRef())
		}
		ans = middleware.Response{
			Err: &util.ErrForbidden,
		}
		return
	}

	var deadline *time.Time
	if formInfo.Deadline.Valid {
		deadline = &formInfo.Deadline.Time
	}

	req := dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: dto.PNCanEnroll,
		Target:   dto.IdentifierString(dto.PTInstitution, ownerInfo.GetOwner()),
		Condition: &dto.RelationCondition{
			Name: "enrollment_available",
			Context: []dto.ContextEntry{
				dto.HavingEntry("institution_verified", dto.CETBool, lookup.Verified),
				dto.HavingEntry("current_time", dto.CETTimestamp, time.Now().String()),
				dto.HavingEntry("deadline", dto.CETTimestamp, deadline),
			},
		},
	}
	res, err := permissions.CheckPermissionInternal(request.Context(), req)
	if err != nil {
		rlog.Error(util.MsgCallError, "err", err)
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}
	if !res.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	return
}

// Validates a user's permission to create an enrollment form
//
//encore:middleware target=tag:can_create_enrollment_form
func AllowedToCreateEnrollmentForm(request middleware.Request, next middleware.Next) (ans middleware.Response) {
	ans = next(request)

	uid, _ := auth.UserID()
	ownerInfo, ok := request.Data().Payload.(models.OwnerInfo)
	if !ok {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	req := dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: "can_create_enrollment_forms",
		Target:   dto.IdentifierString(dto.PTInstitution, ownerInfo.GetOwner()),
	}
	res, err := permissions.CheckPermissionInternal(request.Context(), req)
	if err != nil {
		rlog.Error(util.MsgCallError, "err", err)
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}
	if !res.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	return
}

// Validates a user's permission to create an academic year
//
//encore:middleware target=tag:can_create_academic_year
func AllowedToCreateAcademicYear(request middleware.Request, next middleware.Next) middleware.Response {
	uid, _ := auth.UserID()
	ownerInfo, ok := request.Data().Payload.(models.OwnerInfo)
	if !ok {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	req := dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, uid),
		Relation: dto.PNCanCreateAcademicYear,
		Target:   dto.IdentifierString(dto.PTInstitution, ownerInfo.GetOwner()),
	}
	res, err := permissions.CheckPermissionInternal(request.Context(), req)
	if err != nil {
		rlog.Error(util.MsgCallError, "err", err)
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}
	if !res.Allowed {
		return middleware.Response{
			Err: &util.ErrForbidden,
		}
	}

	var existingCount uint
	if err := db.QueryRow(request.Context(), "SELECT COUNT(institution_id) FROM func_get_last_academic_year($1) a WHERE NOW()::DATE BETWEEN a.start_date AND a.end_date;", ownerInfo.GetOwner()).Scan(&existingCount); err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return middleware.Response{
			Err: &util.ErrUnknown,
		}
	}
	if existingCount > 0 {
		return middleware.Response{
			Err: &errs.Error{
				Code:    errs.FailedPrecondition,
				Message: "There is already an existing academic year",
			},
		}
	}
	return next(request)
}

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

	ans, err := permissions.CheckPermissionInternal(req.Context(), dto.InternalRelationCheckRequest{
		Actor:    dto.IdentifierString(dto.PTUser, userId),
		Relation: dto.PNCanCreateInstitution,
		Target:   dto.IdentifierString(dto.PTTenant, data.TenantId),
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
