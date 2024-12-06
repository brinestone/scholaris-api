package settings_test

import (
	"context"
	crypto "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"math/rand"
	"slices"
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
	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, req dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTForm: {},
			},
		}, nil
	})
}

func TestMain(t *testing.M) {
	mockEndpoints()
	uid, data := makeUser()
	mainContext = auth.WithContext(context.TODO(), uid, &data)
	t.Run()
}

func makeUser() (auth.UID, dto.AuthClaims) {
	uid := uint64(rand.Int63n(1000))
	userData := dto.AuthClaims{
		Email:      gofakeit.Person().Contact.Email,
		Avatar:     &gofakeit.Person().Image,
		FullName:   gofakeit.Name(),
		Provider:   gofakeit.RandomString(sAuth.ValidProviders),
		ExternalId: randomString(30),
		Account:    uint64(gofakeit.UintRange(1, 10000)),
		Sub:        uid,
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
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestUpdateSettings(t *testing.T) {
	owner, ownerType := randomOwner()

	t.Run("EmptyUpdates", func(t *testing.T) {
		testUpdateSettingsWithNoUpdates(t, mainContext, owner, ownerType)
	})

	passed := t.Run("NewSettings", func(t *testing.T) {
		testUpdateSettingsWithNewSettings(t, mainContext, owner, ownerType)
		mockEndpoints()
	})

	if passed {
		et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
			return &dto.ListObjectsResponse{
				Relations: map[dto.PermissionType][]uint64{
					dto.PTSetting: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				},
			}, nil
		})
		res, err := settings.FindSettings(mainContext, dto.GetSettingsRequest{
			Owner:     owner,
			OwnerType: ownerType,
		})
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("ExistingSetting", func(t *testing.T) {
			index := gofakeit.IntRange(0, len(res.Settings)-1)
			if index < 0 || index >= len(res.Settings) {
				t.FailNow()
				return
			}
			setting := slices.Collect(maps.Values(res.Settings))[index]

			testUpdateUsingExistingSetting(t, mainContext, owner, ownerType, setting.Key)
			mockEndpoints()
		})
	}
}

func testUpdateUsingExistingSetting(t *testing.T, ctx context.Context, owner uint64, ownerType, key string) {
	updates := makeUpdates(1)
	updates[0].Key = key
	req := dto.UpdateSettingsRequest{
		OwnerType:    ownerType,
		CaptchaToken: randomString(30),
		Owner:        owner,
		Updates:      updates,
	}

	err := settings.UpdateSettings(ctx, req)
	assert.Nil(t, err)
	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTSetting: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
		}, nil
	})

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

	var target *dto.Setting
	for _, v := range res.Settings {
		if v.Key == key {
			target = &v
			break
		}
	}

	assert.NotNil(t, target)
	assert.Equal(t, key, target.Key)
	assert.Equal(t, owner, target.Owner)
	assert.Equal(t, ownerType, target.OwnerType)
	assert.True(t, target.UpdatedAt.After(target.CreatedAt))
}

func makeUpdates(cnt int) []dto.SettingUpdate {
	var ans []dto.SettingUpdate
	for i := 0; i < cnt; i++ {
		update := dto.SettingUpdate{
			Key:         gofakeit.UUID(),
			Label:       gofakeit.LoremIpsumSentence(5),
			Description: nil,
			MultiValues: gofakeit.Bool(),
			// SystemGenerated: gofakeit.Bool(),
			Parent:      nil,
			Overridable: gofakeit.Bool(),
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
	maxCnt := gofakeit.IntRange(1, 15)
	req := dto.UpdateSettingsRequest{
		OwnerType:    ownerType,
		CaptchaToken: randomString(20),
		Owner:        owner,
		Updates:      makeUpdates(maxCnt),
	}

	err := settings.UpdateSettings(ctx, req)
	assert.Nil(t, err)

	et.MockEndpoint(permissions.ListRelationsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (*dto.ListObjectsResponse, error) {
		var settingIds []uint64
		var i = 0
		for i < maxCnt {
			settingIds = append(settingIds, uint64(i+1))
			i++
		}
		return &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{
				dto.PTSetting: settingIds,
			},
		}, nil
	})

	res, err := settings.FindSettings(mainContext, dto.GetSettingsRequest{
		Owner:     owner,
		OwnerType: ownerType,
	})
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, res)
	assert.True(t, assert.LessOrEqual(t, len(res.Settings), maxCnt) && assert.GreaterOrEqual(t, len(res.Settings), 1))
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
