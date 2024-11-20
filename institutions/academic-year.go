package institutions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/settings"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
)

// Get Academic years
//
//encore:api auth method=POST path=/academic-years
func GetAcademicYears(ctx context.Context, req dto.GetAcademicYearsRequest) (*dto.GetAcademicYearsResponse, error) {
	mods, err := findAcademicYears(ctx, req.GetOwnerId(), req.Page, req.Size)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return nil, &util.ErrUnknown
	}
	return &dto.GetAcademicYearsResponse{AcademicYears: academicYearsToDto(mods...)}, nil
}

// Manually creates an academic year
//
//encore:api auth method=POST path=/academic-years/new tag:can_create_academic_year
func CreateAcademicYear(ctx context.Context, req dto.NewAcademicYearRequest) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return &util.ErrUnknown
	}

	yearId, termIds, err := createAcademicYear(ctx, tx, req.Institution, req.StartOffset, req.TermDurations, req.Vacations)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		return &util.ErrUnknown
	}

	r := dto.UpdatePermissionsRequest{
		Updates: []dto.PermissionUpdate{
			{
				Actor:    dto.IdentifierString(dto.PTInstitution, req.GetOwner()),
				Relation: models.PermOwner,
				Target:   dto.IdentifierString(dto.PTAcademicYear, yearId),
			},
		},
	}

	for _, termId := range termIds {
		r.Updates = append(r.Updates, dto.NewPermissionUpdate[uint64](dto.IdentifierString(dto.PTAcademicYear, yearId), models.PermOwner, dto.IdentifierString(dto.PTAcademicTerm, termId)))
	}

	if err := permissions.SetPermissions(ctx, r); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgCallError, "err", err)
		return &util.ErrUnknown
	}

	if err := tx.Commit(); err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return &util.ErrUnknown
	}
	return nil
}

// Automatically creates academic years
//
//encore:api private method=POST path=/academic-years/new/auto
func AutoCreateAcademicYears(ctx context.Context) error {
	stats, err := getInstitutionStats(ctx)
	if err != nil {
		return err
	}

	var size int = 1000
	var wg sync.WaitGroup
	for page := 0; page*size < int(stats.TotalVerified); page++ {
		wg.Add(1)
		go doAutocreateAcademicYears(ctx, &wg, page, size)
	}

	wg.Wait()
	return nil
}

func doAutocreateAcademicYears(ctx context.Context, wg *sync.WaitGroup, page, size int) {
	defer wg.Done()

	query := `
		SELECT
			id
		FROM
			institutions
		WHERE
			verified=true
		OFFSET $1
		LIMIT $2
		;
	`
	rows, err := db.Query(ctx, query, page*size, size)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		tx, err := db.Begin(ctx)
		if err != nil {
			rlog.Debug(util.MsgDbAccessError, "err", err)
			return
		}

		var institution uint64
		if err := rows.Scan(&institution); err != nil {
			tx.Rollback()
			rlog.Debug(util.MsgDbAccessError, "err", err)
			return
		}

		yearAndTermsMap, err := autoCreateAcademicYear(ctx, tx, institution)
		if err != nil {
			tx.Rollback()
			rlog.Debug("error while auto-creating academic year", "institution", institution, "err", err)
			return
		}

		if len(yearAndTermsMap) == 0 {
			rlog.Warn("skipped academic year auto-creation", "institution", institution)
		}

		for yearId, termIds := range yearAndTermsMap {
			req := dto.UpdatePermissionsRequest{
				Updates: []dto.PermissionUpdate{
					{
						Actor:    dto.IdentifierString(dto.PTInstitution, institution),
						Relation: models.PermOwner,
						Target:   dto.IdentifierString(dto.PTAcademicYear, yearId),
					},
				},
			}

			for _, termId := range termIds {
				req.Updates = append(req.Updates, dto.NewPermissionUpdate[uint64](dto.IdentifierString(dto.PTAcademicYear, yearId), models.PermOwner, dto.IdentifierString(dto.PTAcademicTerm, termId)))
			}

			if err := permissions.SetPermissions(ctx, req); err != nil {
				tx.Rollback()
				rlog.Error(util.MsgCallError, "err", err)
				return
			}

			if err := tx.Commit(); err != nil {
				rlog.Error(util.MsgDbAccessError, "err", err)
				return
			}
		}
	}
}

// Creates an institution's academic year.
func autoCreateAcademicYear(ctx context.Context, tx *sqldb.Tx, institution uint64) (ans map[uint64][]uint64, err error) {
	ans = make(map[uint64][]uint64)
	res, err := settings.FindSettingsInternal(ctx, dto.GetSettingsInternalRequest{
		Owner: institution,
	})
	if err != nil {
		return
	}

	for _, s := range res.Settings {
		switch s.Key {
		case dto.SKAcademicYearAutoCreation:
			valPtr := s.Values[0].Value
			if valPtr == nil {
				break
			}
			autoAcademicYearCreationEnabled, err := strconv.ParseBool(*valPtr)
			if err != nil {
				rlog.Error("parse error", "err", err)
			}

			if err != nil || !autoAcademicYearCreationEnabled {
				break
			}

			var creationOffset, startOffset string
			var academicYearDuration string
			var academicTermDurations, vacationDurations []string

			for _, ss := range res.Settings {
				switch ss.Key {
				case dto.SKAcademicYearAutoCreationOffset:
					if len(ss.Values) > 0 && ss.Values[0].Value != nil {
						creationOffset = *ss.Values[0].Value
					} else {
						return nil, fmt.Errorf("required setting not set - [%s]", ss.Key)
					}
				case dto.SKAcademicYearStartDateOffset:
					if len(ss.Values) > 0 && ss.Values[0].Value != nil {
						startOffset = *ss.Values[0].Value
					} else {
						return nil, fmt.Errorf("required setting not set - [%s]", ss.Key)
					}
				case dto.SKAcademicYearDuration:
					if len(ss.Values) > 0 && ss.Values[0].Value != nil {
						academicYearDuration = *ss.Values[0].Value
					} else {
						rlog.Error("required setting not set", "key", ss.Key)
					}
				case dto.SKAcademicTermDurations:
					if len(ss.Values) > 0 {
						for _, ssv := range ss.Values {
							if ssv.Value == nil || len(*ssv.Value) == 0 {
								continue
							}
							academicTermDurations = append(academicTermDurations, *ssv.Value)
						}
					} else {
						rlog.Error("required setting not set", "key", ss.Key)
					}
				case dto.SKVacationDurations:
					if len(ss.Values) > 0 {
						for _, ssv := range ss.Values {
							if ssv.Value == nil || len(*ssv.Value) == 0 {
								continue
							}
							vacationDurations = append(vacationDurations, *ssv.Value)
						}
					} else {
						rlog.Error("required setting not set", "key", ss.Key)
					}
				}
			}

			if len(creationOffset) == 0 || len(startOffset) == 0 || len(academicYearDuration) == 0 || len(academicTermDurations) == 0 {
				return nil, errors.New("insuficient parameters for auto academic year creation")
			}

			var yearId uint64
			var termIds []uint64
			if err := tx.QueryRow(ctx, "SELECT year_id, term_ids FROM func_auto_create_academic_year($1,$2,$3,$4,$5,$6);", creationOffset, startOffset, institution, academicYearDuration, pq.Array(academicTermDurations), pq.Array(vacationDurations)).Scan(&yearId, pq.Array(&termIds)); err != nil {
				return nil, err
			}
			ans[yearId] = termIds
		}
	}
	return ans, nil
}

func createAcademicYear(ctx context.Context, tx *sqldb.Tx, institution uint64, startOffset time.Duration, academicTermDurations, vacationDurations []time.Duration) (yearId uint64, termIds []uint64, err error) {
	startOffsetString := fmt.Sprintf("%.0f hours", startOffset.Abs().Hours())
	vacationDurationStrings := make([]string, len(vacationDurations))
	termDurationStrings := make([]string, len(academicTermDurations))

	for i, v := range vacationDurations {
		vacationDurationStrings[i] = fmt.Sprintf("%.0f hours", v.Abs().Hours())
	}

	for i, v := range academicTermDurations {
		termDurationStrings[i] = fmt.Sprintf("%.0f hours", v.Abs().Hours())
	}

	if err = tx.QueryRow(ctx, "SELECT year_id FROM func_create_academic_year($1,$2,$3,$4);", institution, startOffsetString, pq.Array(termDurationStrings), pq.Array(vacationDurationStrings)).Scan(&yearId); err != nil {
		return
	}
	return
}

func findAcademicYears(ctx context.Context, institution uint64, page, size uint) (ans []*models.AcademicYear, err error) {
	query := `
		SELECT 
			*
		FROM 
			vw_AllAcademicYears 
		WHERE 
			institution_id=$3
		OFFSET $1 LIMIT $2;
	`

	var defaultSize uint = 100
	actualSize := size
	if size <= 0 {
		actualSize = defaultSize
	}
	rows, err := db.Query(ctx, query, page*actualSize, actualSize, institution)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	yearMaps := make(map[uint64]*[]models.AcademicTerm)
	var yearIds []uint64
	for rows.Next() {
		ay := new(models.AcademicYear)
		if err = rows.Scan(&ay.Id, &ay.Institution, &ay.StartDate, &ay.Duration, &ay.EndDate, &ay.Label, &ay.CreatedAt, &ay.UpdatedAt); err != nil {
			return
		}

		ans = append(ans, ay)
		yearMaps[ay.Id] = &ay.Terms
		yearIds = append(yearIds, ay.Id)
	}

	query = `
			SELECT
				*
			FROM
				vw_AllAcademicTerms
			WHERE
				institution=$1 AND year_id=ANY($2)
			;
		`
	rows2, err := db.Query(ctx, query, institution, pq.Array(yearIds))
	if !errors.Is(err, sqldb.ErrNoRows) && err != nil {
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var at models.AcademicTerm
		if err := rows2.Scan(&at.Id, &at.Year, &at.Institution, &at.StartDate, &at.Duration, &at.EndDate, &at.Label, &at.CreatedAt, &at.UpdatedAt); err != nil {
			return nil, err
		}
		termsRef := *yearMaps[at.Year]
		*yearMaps[at.Year] = append(termsRef, at)
	}

	return
}

func academicYearsToDto(year ...*models.AcademicYear) []dto.AcademicYear {
	var ans []dto.AcademicYear = make([]dto.AcademicYear, len(year))

	for i, v := range year {
		var _dto dto.AcademicYear = dto.AcademicYear{
			Id:          v.Id,
			Institution: v.Institution,
			StartDate:   v.StartDate,
			Duration:    v.Duration,
			EndDate:     v.EndDate,
			Label:       v.Label,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			Terms:       make([]dto.AcademicTerm, len(v.Terms)),
		}

		for j, v := range v.Terms {
			_dto.Terms[j] = dto.AcademicTerm{
				Id:        v.Id,
				Year:      v.Year,
				StartDate: v.StartDate,
				CreatedAt: v.CreatedAt,
				Duration:  v.Duration,
				EndDate:   v.EndDate,
				Label:     v.Label,
				UpdatedAt: v.UpdatedAt,
			}
		}

		ans[i] = _dto
	}

	return ans
}
