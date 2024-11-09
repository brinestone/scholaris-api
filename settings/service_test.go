package settings_test

import (
	"context"
	crypto "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"encore.dev/beta/auth"
	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	sAuth "github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/settings"
	"github.com/stretchr/testify/assert"
)

var mainContext context.Context

func TestMain(t *testing.M) {
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

func randomOwner() (uint64, string) {
	return uint64(gofakeit.UintRange(1, 5000)), gofakeit.RandomString([]string{string(dto.PTInstitution), string(dto.PTTenant)})
}

func randomString(len int) string {
	buf := make([]byte, len)
	io.ReadFull(crypto.Reader, buf)
	return hex.EncodeToString(buf)
}

func TestGetSettings(t *testing.T) {
	ownerId, ownerType := randomOwner()

	res, err := settings.FindSettings(mainContext, dto.GetSettingsRequest{
		Owner:     ownerId,
		OwnerType: ownerType,
	})
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
}

func TestUpdateSettings(t *testing.T) {
	owner, ownerType := randomOwner()

	t.Run("EmptyUpdates", func(t *testing.T) {
		testUpdateSettingsWithNoUpdates(t, mainContext, owner, ownerType)
	})

	t.Run("NewSettings", func(t *testing.T) {
		testUpdateSettingsWithNewSettings(t, mainContext, owner, ownerType)
	})
}

func makeUpdates() []dto.SettingUpdate {
	cnt := gofakeit.IntRange(1, 20)
	var ans []dto.SettingUpdate
	for i := 0; i < cnt; i++ {
		key := gofakeit.UUID()
		update := dto.SettingUpdate{
			Key:             &key,
			Label:           gofakeit.LoremIpsumSentence(5),
			Description:     nil,
			MultiValues:     gofakeit.Bool(),
			SystemGenerated: gofakeit.Bool(),
			Parent:          nil,
			Overridable:     gofakeit.Bool(),
		}
		if update.MultiValues {
			cnt2 := gofakeit.IntRange(1, 5)
			for j := 0; j < cnt2; j++ {
				var val *string
				if gofakeit.Bool() {
					tmp := gofakeit.LoremIpsumSentence(gofakeit.IntRange(1, 3))
					val = &tmp
				}
				option := dto.SettingOptionUpdate{
					Label: gofakeit.LoremIpsumSentence(gofakeit.IntRange(3, 8)),
					Value: val,
				}
				update.Options = append(update.Options, option)
			}
		}
		ans = append(ans, update)
	}
	return ans
}

func testUpdateSettingsWithNewSettings(t *testing.T, ctx context.Context, owner uint64, ownerType string) {
	req := dto.UpdateSettingsRequest{
		OwnerType:    ownerType,
		CaptchaToken: randomString(20),
		Owner:        owner,
		Updates:      makeUpdates(),
	}

	err := settings.UpdateSettings(ctx, req)
	assert.Nil(t, err)

	res, err := settings.FindSettings(mainContext, dto.GetSettingsRequest{
		Owner:     owner,
		OwnerType: ownerType,
	})
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Settings)
}

func testUpdateSettingsWithNoUpdates(t *testing.T, ctx context.Context, owner uint64, ownerType string) {
	req := dto.UpdateSettingsRequest{
		OwnerType:    ownerType,
		CaptchaToken: randomString(20),
		Owner:        owner,
	}
	err := settings.UpdateSettings(ctx, req)
	assert.NotNil(t, err)
}
