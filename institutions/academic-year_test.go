package institutions_test

import (
	"context"
	crypto "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/institutions"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/tenants"
	"github.com/stretchr/testify/assert"
)

var (
	mainContext    context.Context
	tenantId       uint64
	institutionRef *dto.InstitutionLookup
)

func mockEndpoints() {
	et.MockEndpoint(permissions.CheckPermission, func(ctx context.Context, req dto.RelationCheckRequest) (*dto.RelationCheckResponse, error) {
		return &dto.RelationCheckResponse{
			Allowed: true,
		}, nil
	})
	et.MockEndpoint(permissions.SetPermissions, func(ctx context.Context, req dto.UpdatePermissionsRequest) error {
		return nil
	})
	et.MockEndpoint(sAuth.VerifyCaptchaToken, func(ctx context.Context, req sAuth.VerifyCaptchaRequest) error {
		return nil
	})
	et.MockEndpoint(permissions.ListRelations, func(ctx context.Context, req dto.ListRelationsRequest) (*dto.ListRelationsResponse, error) {
		return &dto.ListRelationsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTForm: {},
			},
		}, nil
	})
	et.MockEndpoint(tenants.FindTenant, func(ctx context.Context, id uint64) (*models.Tenant, error) {
		date := gofakeit.PastDate()
		return &models.Tenant{
			Name:         gofakeit.LoremIpsumSentence(gofakeit.IntRange(1, 5)),
			Id:           id,
			CreatedAt:    date,
			UpdatedAt:    date,
			Subscription: gofakeit.Uint64(),
		}, nil
	})
}

func TestMain(t *testing.M) {
	mockEndpoints()
	uid, data := makeUser()
	mainContext = auth.WithContext(context.TODO(), uid, &data)
	t.Run()
}

func makeUser() (auth.UID, sAuth.AuthClaims) {
	uid := uint64(rand.Int63n(1000))
	userData := sAuth.AuthClaims{
		Email:    gofakeit.Person().Contact.Email,
		Avatar:   gofakeit.Person().Image,
		FullName: gofakeit.Name(),
		Sub:      uid,
	}
	return auth.UID(fmt.Sprintf("%d", uid)), userData
}

////func randomOwner() (uint64, string) {
////	return uint64(gofakeit.UintRange(1, 5000)), gofakeit.RandomString([]string{string(dto.PTInstitution), string(dto.PTTenant)})
////}

func randomString(len int) string {
	buf := make([]byte, len)
	io.ReadFull(crypto.Reader, buf)
	return hex.EncodeToString(buf)
}

func assertTenantId() uint64 {
	if tenantId == 0 {
		tenantId = uint64(gofakeit.UintRange(1, 50000))
	}
	return tenantId
}

func makeInstitution() (*dto.InstitutionLookup, error) {

	request := dto.NewInstitutionRequest{
		Name:        gofakeit.Product().Name,
		Description: gofakeit.LoremIpsumParagraph(1, gofakeit.IntRange(0, 3), gofakeit.IntRange(0, 5), "\n\t"),
		Logo:        gofakeit.ImageURL(400, 400),
		Slug:        gofakeit.UUID()[:15],
		TenantId:    assertTenantId(),
		Captcha:     randomString(30),
		Timestamp:   time.Now(),
	}

	info, err := institutions.NewInstitution(mainContext, request)
	if err != nil {
		return nil, err
	}

	institutionRef = info
	return info, nil
}

func TestCreateAcademicYear(t *testing.T) {
	institution, err := makeInstitution()
	if err != nil {
		t.Error(err)
		return
	}

	termCount := 3
	vacationCount := termCount - 1
	yearDuration := time.Hour * 6570
	vacationDuration := time.Hour * 336
	termDuration := yearDuration / time.Duration(termCount)

	request := dto.NewAcademicYearRequest{
		Institution: institution.Id,
		StartOffset: time.Hour * 2,
	}

	for i := 0; i < vacationCount; i++ {
		request.Vacations = append(request.Vacations, vacationDuration)
	}

	for i := 0; i < termCount; i++ {
		request.TermDurations = append(request.TermDurations, termDuration)
	}

	if err := institutions.CreateAcademicYear(mainContext, request); err != nil {
		t.Error(err)
	}
}

func TestGetAcademicYears(t *testing.T) {
	if institutionRef == nil {
		t.Skip()
	}

	res, err := institutions.GetAcademicYears(mainContext, dto.GetAcademicYearsRequest{
		Institution: institutionRef.Id,
		Size:        4,
	})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.AcademicYears)
	for _, v := range res.AcademicYears {
		assert.Equal(t, v.Duration, v.EndDate.Sub(v.StartDate))
		assert.NotEmpty(t, v.Terms)
	}
}
