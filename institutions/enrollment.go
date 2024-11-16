package institutions

import (
	"context"
	"strconv"
	"strings"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/forms"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/settings"
	"github.com/brinestone/scholaris/util"
)

// Creates an enrollment
//
//encore:api auth method=POST path=/institutions/enroll tag:can_enroll
func NewEnrollment(ctx context.Context, req dto.NewEnrollmentRequest) (err error) {
	userId, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(userId), 10, 64)

	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Debug(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	formInfo, _ := findLevelEnrollmentForm(ctx, req.GetLevelRef(), req.GetOwner())

	res, err := forms.CreateFormResponseInternal(ctx, formInfo.Id)
	if err != nil {
		return
	}

	if err = createEnrollment(ctx, tx, res.Id, formInfo.Id, uid, req.GetOwner()); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}
	tx.Commit()

	return
}

// Creates an enrollment form
//
//encore:api auth method=POST path=/institutions/enrollment-forms tag:can_create_enrollment_form tag:needs_captcha_ver
func NewEnrollmentForm(ctx context.Context, req dto.NewEnrollmentFormRequest) (err error) {
	settings, err := settings.FindSettings(ctx, dto.GetSettingsRequest{
		Owner:     req.GetOwner(),
		OwnerType: req.GetOwnerType(),
	})

	if err != nil {
		rlog.Error(util.MsgCallError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	var responseWindowSetting dto.Setting
	var responseWindow *time.Duration
	responseWindowSetting, ok := settings.Settings[dto.SKDefaultEnrollmentResponseWindow]
	if ok {
		responseWindowStr := *responseWindowSetting.Values[0].Value
		d, _ := strconv.ParseInt(strings.Split(responseWindowStr, " ")[1], 10, 64)
		tmp := time.Hour * time.Duration(d)
		responseWindow = &tmp
	}

	formInfo, err := forms.NewForm(ctx, dto.NewFormInput{
		Title:           "Untitled Form",
		BackgroundColor: nil,
		MultiResponse:   false,
		Resubmission:    false,
		CaptchaToken:    req.GetCaptchaToken(),
		Owner:           req.GetOwner(),
		OwnerType:       req.GetOwnerType(),
		ResponseWindow:  responseWindow,
		MaxResponses:    1,
		MaxSubmissions:  1,
		Tags:            []string{"enrollment"},
	})

	if err != nil {
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	if err = registerEnrollmentForm(ctx, tx, formInfo.Id, req.GetOwner()); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		err = &util.ErrUnknown
		return
	}

	tx.Commit()
	return
}

func registerEnrollmentForm(ctx context.Context, tx *sqldb.Tx, form, institution uint64) (err error) {
	query := `
		INSERT INTO
			enrollment_forms(form,institution)
		VALUES
			($1,$2);
	`

	_, err = tx.Exec(ctx, query, form, institution)
	return
}

func findLevelEnrollmentForm(ctx context.Context, level, institution uint64) (ans *models.Form, err error) {
	query := `
		SELECT
			form
		FROM
			enrollment_forms
		WHERE
			institution=$1 AND level=$2
		;
	`
	var formId uint64
	if err = db.QueryRow(ctx, query, institution, level).Scan(&formId); err != nil {
		return
	}

	ans, err = forms.GetFormInfoInternal(ctx, formId)
	return
}

func createEnrollment(ctx context.Context, tx *sqldb.Tx, response, form, user, institution uint64) (err error) {
	query := `
		INSERT INTO
			enrollments(id,form,institution,responder)
		VALUES
			($1,$2,$3);
	`

	_, err = tx.Exec(ctx, query, response, form, institution, user)
	return
}
