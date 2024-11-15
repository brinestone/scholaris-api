package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"encore.dev/beta/errs"
)

// Well-known institution setting keys
const (
	SKAcademicYearDuration            = "academicYearDuration"
	SKAcademicTermCount               = "academicTermsCount"
	SKDefaultEnrollmentResponseWindow = "defaultEnrollmentResponseWindow"
	SKAcademicYearStartDateOffset     = "academicYearStartDateOffset"
	SKAcademicTermDurations           = "academicTermDurations"
	SKVacationDurations               = "vacationDurations"
	SKAcademicYearAutoCreation        = "academicYearAutoCreation"
	SKAcademicYearAutoCreationOffset  = "academicYearAutoCreationOffset"
)

type AcademicTerm struct {
	Id        uint64        `json:"id"`
	Year      uint64        `json:"academicYear"`
	StartDate time.Time     `json:"startDate"`
	CreatedAt time.Time     `json:"createdAt"`
	Duration  time.Duration `json:"duration"`
	EndDate   time.Time     `json:"endDate"`
	Label     string        `json:"label"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type AcademicYear struct {
	Id          uint64         `json:"id"`
	Institution uint64         `json:"institution"`
	StartDate   time.Time      `json:"startDate"`
	Duration    time.Duration  `json:"duration"`
	EndDate     time.Time      `json:"endDate"`
	Label       string         `json:"label"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Terms       []AcademicTerm `json:"academicTerms"`
}

type GetAcademicYearsResponse struct {
	AcademicYears []AcademicYear `json:"academicYears"`
}

type GetAcademicYearsRequest struct {
	Institution uint64 `header:"x-institution"`
	Page        uint   `query:"page" encore:"optional"`
	Size        uint   `query:"size" encore:"optional"`
}

func (g GetAcademicYearsRequest) GetOwnerId() uint64 {
	return g.Institution
}

func (g GetAcademicYearsRequest) GetOwnerType() string {
	return string(PTInstitution)
}

// Data for creation of an academic year.
type NewAcademicYearRequest struct {
	// The institution owning the academic year.
	Institution uint64 `header:"x-institution"`
	// The time between the creation and official launching of the academic year being created.
	StartOffset time.Duration `json:"startOffset"`
	// The durations of break periods in the academic year being created.
	Vacations []time.Duration `json:"vacations"`
	// The durations of the terms of the academic year being created.
	TermDurations []time.Duration `json:"termDurations"`
}

func (n NewAcademicYearRequest) GetOwnerType() string {
	return string(PTInstitution)
}

func (n NewAcademicYearRequest) GetOwner() uint64 {
	return n.Institution
}

func (n NewAcademicYearRequest) Validate() error {
	msgs := make([]string, 0)
	if n.Institution == 0 {
		msgs = append(msgs, "The x-institution header is required")
	}

	if len(n.TermDurations) == 0 {
		msgs = append(msgs, "The termDurations field cannot be empty")
	} else if len(n.Vacations) > 0 && len(n.Vacations) != len(n.TermDurations)-1 {
		msgs = append(msgs, fmt.Sprintf("The vacations field should either be empty or have a length of %d. (i.e len(termDurations)-1) ", len(n.TermDurations)-1))
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

type Institution struct {
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty" encore:"optional"`
	Logo        *string   `json:"logo,omitempty" encore:"optional"`
	Visible     bool      `json:"visible"`
	Slug        string    `json:"slug"`
	Id          uint64    `json:"id,omitempty" encore:"optional"`
	TenantId    uint64    `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	IsMember    bool      `json:"isMember"`
	Members     int       `json:"members"`
}

type InstitutionLookup struct {
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty" encore:"optional"`
	Logo        *string   `json:"logo,omitempty" encore:"optional"`
	Visible     bool      `json:"visible"`
	Slug        string    `json:"slug"`
	Id          uint64    `json:"id,omitempty" encore:"optional"`
	TenantId    uint64    `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	IsMember    bool      `json:"isMember"`
}

type NewInstitutionRequest struct {
	// The institution's name
	Name string `json:"name"`
	// The institution's description (optional)
	Description string `json:"description,omitempty" encore:"optional"`
	// The institution's logo (optional)
	Logo string `json:"logo,omitempty" encore:"optional"`
	// The institution's slug
	Slug string `json:"slug"`
	// The institution's tenant ID
	TenantId uint64 `json:"tenantId"`
	// The request's captcha token
	Captcha string `json:"captcha"`
	// The timestamp of the request
	Timestamp time.Time `header:"x-timestamp"`
}

func (n NewInstitutionRequest) GetCaptchaToken() string {
	return n.Captcha
}

func (n NewInstitutionRequest) Validate() error {
	msgs := make([]string, 0)

	if n.TenantId == 0 {
		msgs = append(msgs, "The tenantId field is required")
	}

	if len(n.Name) == 0 {
		msgs = append(msgs, "The name field is required")
	}

	if len(n.Slug) == 0 {
		msgs = append(msgs, "The slug field is required")
	}

	if len(msgs) > 0 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: strings.Join(msgs, "\n"),
		}
	}

	return nil
}
