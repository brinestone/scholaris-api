package institutions

import (
	"context"
	"strconv"
	"strings"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/forms"
	"github.com/brinestone/scholaris/settings"
	"github.com/brinestone/scholaris/util"
)

// Creates an enrollment
//
//encore:api auth method=POST path=/institutions/enroll
func NewEnrollment(ctx context.Context, req dto.NewEnrollmentRequest) (err error) {

	return
}

// Creates an enrollment form
//
//encore:api auth method=POST path=/institutions/enrollment-forms tag:can_create_enrollment_form
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

	var deadlineSetting dto.Setting
	var deadline *time.Time
	deadlineSetting, ok := settings.Settings[dto.SKDefaultEnrollmentResponseWindow]
	if ok {
		deadlineStr := *deadlineSetting.Values[0].Value
		d, _ := strconv.ParseInt(strings.Split(deadlineStr, " ")[1], 10, 64)
		tmp := time.Now().Add(time.Hour * time.Duration(d))
		deadline = &tmp
	}

	formInfo, err := forms.NewForm(ctx, dto.NewFormInput{
		Title:           "Untitled Form",
		BackgroundColor: nil,
		MultiResponse:   false,
		Resubmission:    false,
		CaptchaToken:    req.GetCaptchaToken(),
		Owner:           req.GetOwner(),
		OwnerType:       req.GetOwnerType(),
		Deadline:        deadline,
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
