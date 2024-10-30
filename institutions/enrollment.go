package institutions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/billing"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Get Institution's Enrollment questions
//
//encore:api public method=GET path=/institutions/:id/enrollment-questions
func GetEnrollmentQuestions(ctx context.Context, id uint64) (*dto.EnrollmentQuestions, error) {
	models, err := findEnrollmentQuestions(ctx, id)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &util.ErrNotFound
	} else if err != nil {
		rlog.Error(err.Error())
		if errs.Convert(err) != nil {
			return nil, err
		} else {
			return nil, &util.ErrUnknown
		}
	}

	var dtos = make([]*dto.EnrollmentQuestion, len(models))
	for i, model := range models {
		dtos[i] = enrollmentQuestionToDto(model)
		enrollmentQuestionCache.Set(ctx, model.Id, *dtos[i])
	}

	return &dto.EnrollmentQuestions{
		Questions: dtos,
	}, nil
}

// Creates a new enrollment
//
//encore:api auth method=POST path=/institutions/enroll
func NewEnrollment(ctx context.Context, input dto.NewEnrollment) (*dto.EnrollmentState, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	// Create the enrollment
	owner, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(owner), 10, 64)
	enrollment, err := createEnrollment(ctx, tx, input, uid)
	if err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return nil, err
		}
		rlog.Error("error while creating enrollment", "msg", err.Error())
		return nil, &util.ErrUnknown
	}

	if err := permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Target:   dto.IdentifierString(dto.PTEnrollment, enrollment.Id),
				Subject:  dto.IdentifierString(dto.PTUser, owner),
				Relation: models.PermOwner,
			},
			{
				Target:   dto.IdentifierString(dto.PTEnrollment, enrollment.Id),
				Subject:  dto.IdentifierString(dto.PTInstitution, input.Destination),
				Relation: models.PermDestination,
			},
		},
	}); err != nil {
		rlog.Error("error while updating permissions", "msg", err.Error())
		tx.Rollback()
		return nil, &util.ErrUnknown
	}
	defer tx.Commit()

	return enrollmentToDto(enrollment), nil
}

func findEnrollmentQuestions(ctx context.Context, institutionId uint64) ([]*models.EnrollmentFormQuestion, error) {
	query := `
		SELECT
			efq.id,
			efq.institution,
			efq.prompt,
			efq.q_type,
			efq.a_type,
			efq.is_required,
			efq.choice_delimiter,
			COALESCE(json_agg(json_build_object(
				'label', efqo.label,
				'value', efqo.value,
				'isDefault', efqo.is_default
			)) FILTER (WHERE efqo.question IS NOT NULL), '[]') AS answers
		FROM
			enrollment_form_questions efq
		LEFT JOIN
			enrollment_form_question_options efqo ON efq.id = efqo.question
		WHERE
			efq.institution = $1
		GROUP BY
			efq.id
		;
	`
	var ans = make([]*models.EnrollmentFormQuestion, 0)

	rows, err := db.Query(ctx, query, institutionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q = new(models.EnrollmentFormQuestion)
		var optionsJson string
		if err := rows.Scan(&q.Id, &q.Institution, &q.Prompt, &q.QuestionType, &q.AnswerType, &q.IsRequired, &q.ChoiceDelimiter, &optionsJson); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(optionsJson), &q.Options); err != nil {
			return nil, err
		}

		ans = append(ans, q)
	}

	return ans, nil
}

func createEnrollment(ctx context.Context, tx *sqldb.Tx, input dto.NewEnrollment, owner uint64) (*models.Enrollment, error) {

	res, err := billing.VerifyTransaction(ctx, billing.VerifyTransactionRequest{
		VerificationToken: input.ServiceTransactionToken,
	})
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "Service fee transaction verification failed. Please try again later",
		}
	}

	query := `
		SELECT
			COUNT(id)
		FROM
			enrollments
		WHERE
			owner = $1;
	`
	var cnt int
	if err := tx.QueryRow(ctx, query, owner).Scan(&cnt); err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, &util.ErrConflict
	}

	query = `
		INSERT INTO
			enrollments(destination, owner, service_transaction)
		VALUES
			($1,$2,$3)
		RETURNING id;
	`

	var enrollmentId uint64
	if err := tx.QueryRow(ctx, query, input.Destination, owner, res.TransactionId).Scan(&enrollmentId); err != nil {
		return nil, err
	}

	return findEnrollmentByIdFromDbTx(ctx, tx, enrollmentId)
}

func findEnrollmentByIdFromDbTx(ctx context.Context, tx *sqldb.Tx, id uint64) (*models.Enrollment, error) {
	return findEnrollmentByKeyFromDbTx(ctx, tx, "id", id)
}

func findEnrollmentByKeyFromDbTx(ctx context.Context, tx *sqldb.Tx, key string, value any) (*models.Enrollment, error) {
	query := fmt.Sprintf(`
		SELECT
			e.id,
			e.owner,
			e.approved_by,
			e.approved_at,
			e.payment_transaction,
			e.service_transaction,
			e.created_at,
			e.updated_at,
			e.status,
			e.destination,
			COALESCE(json_agg(json_build_object(
				'value', efa.ans
				'answeredAt', efa.answered_at
				'updatedAt', efa.updated_at
				'question', efa.question
			)) FILTER (WHERE efa.id IS NOT NULL), '[]') AS "answers",
			COALESCE(json_agg(ed.url) FILTER (WHERE ed.id IS NOT NULL), '[]') AS "documents"
		FROM
			enrollments e
		LEFT JOIN
			enrollment_documents ed ON ed.enrollment=e.id
		LEFT JOIN
			enrollment_form_answers efa ON efa.enrollment=e.id
		WHERE
			%s=$1
		GROUP BY
			e.id
		;
	`, key)

	var e models.Enrollment
	var answersJson, documentsJson string
	if err := tx.QueryRow(ctx, query, value).
		Scan(&e.Id, &e.Owner, &e.Approver, &e.ApprovedAt, &e.PaymentTransaction, &e.ServiceTransaction, &e.CreatedAt, &e.UpdatedAt, &e.Status, &e.Destination, &answersJson, &documentsJson); err != nil {
		return nil, err
	}

	var answers []models.EnrollmentFormAnswer
	if err := json.Unmarshal([]byte(answersJson), &answers); err != nil {
		return &e, err
	}
	e.Answers = answers

	var documents []string
	if err := json.Unmarshal([]byte(documentsJson), &documents); err != nil {
		return &e, err
	}
	e.Documents = documents

	return &e, nil
}

func enrollmentQuestionToDto(e *models.EnrollmentFormQuestion) *dto.EnrollmentQuestion {
	var ans = dto.EnrollmentQuestion{
		Id:          e.Id,
		Institution: e.Institution,
		Prompt:      e.Prompt,
		Options:     make([]*dto.EnrollmentQuestionOption, len(e.Options)),
		IsRequired:  e.IsRequired.Valid && e.IsRequired.Bool,
	}

	if e.AnswerType.Valid {
		ans.AnswerType = e.AnswerType.String
	}

	if e.QuestionType.Valid {
		ans.QuestionType = e.QuestionType.String
	}

	if e.ChoiceDelimiter.Valid {
		ans.ChoiceDelimiter = e.ChoiceDelimiter.String
	}

	for i, option := range e.Options {
		var o = new(dto.EnrollmentQuestionOption)
		ans.Options[i] = o

		o.IsDefault = option.IsDefault
		o.Label = option.Label
		o.Value = option.Value
	}

	return &ans
}

func enrollmentToDto(e *models.Enrollment) *dto.EnrollmentState {
	var ans = dto.EnrollmentState{
		Id:          e.Id,
		Destination: e.Destination,
		Documents:   e.Documents,
		Answers: make([]struct {
			Value    []*string `json:"value,omitempty"`
			Question uint64    "json:\"question\""
		}, 0),
	}

	for _, a := range e.Answers {
		answer := struct {
			Value    []*string "json:\"value,omitempty\""
			Question uint64    "json:\"question\""
		}{
			Value:    make([]*string, len(a.Value)),
			Question: a.QuestionId,
		}

		if len(a.Value) > 0 {
			for k, i := range a.Value {
				var x *string
				if i.Valid {
					x = &i.String
				}
				answer.Value[k] = x
			}
		}

		ans.Answers = append(ans.Answers, answer)
	}

	return &ans
}
