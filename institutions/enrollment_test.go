package institutions_test

import (
	"context"
	"testing"

	"encore.dev/et"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/institutions"
	"github.com/brinestone/scholaris/settings"
	"github.com/stretchr/testify/assert"
)

func TestNewEnrollment(t *testing.T) {

}

func TestNewEnrollmentForm(t *testing.T) {
	i, err := makeInstitution()
	if err != nil {
		t.Error(err)
		return
	}

	defaultSettings := map[string][]dto.SetValue{}

	et.MockEndpoint(settings.UpdateSettingsInternal, func(ctx context.Context, req dto.UpdateSettingsInternalRequest) error {
		for _, v := range req.Updates {
			defaultSettings[v.Key] = nil
		}
		return nil
	})

	et.MockEndpoint(settings.SetSettingValuesInternal, func(ctx context.Context, req dto.SetSettingValueRequest) error {
		for _, v := range req.Updates {
			defaultSettings[v.Key] = v.Value
		}
		return nil
	})

	et.MockEndpoint(permissions.ListObjectsInternal, func(ctx context.Context, p dto.ListObjectsRequest) (ans *dto.ListObjectsResponse, err error) {
		ans = &dto.ListObjectsResponse{
			Relations: map[dto.PermissionType][]uint64{},
		}

		var i uint64 = 1
		for range defaultSettings {
			ans.Relations[dto.PTInstitution] = append(ans.Relations[dto.PTInstitution], i)
			i++
		}
		return
	})

	et.MockEndpoint(settings.FindSettings, func(ctx context.Context, req dto.GetSettingsRequest) (ans *dto.GetSettingsResponse, err error) {
		ans = &dto.GetSettingsResponse{
			Settings: make(map[string]dto.Setting),
		}

		var i uint64 = 1
		for k, v := range defaultSettings {
			setting := dto.Setting{
				Id:  i,
				Key: k,
			}
			for ii, w := range v {
				setting.Values = append(ans.Settings[k].Values, dto.SettingValue{
					Id:      gofakeit.Uint64(),
					Setting: i,
					SetBy:   gofakeit.Uint64(),
					Index:   uint(ii),
					Value:   w.Value,
				})
			}
			ans.Settings[k] = setting
		}
		return
	})

	err = institutions.NewEnrollmentForm(mainContext, dto.NewEnrollmentFormRequest{
		Institution:  i.Id,
		CaptchaToken: randomString(20),
	})

	assert.Nil(t, err)
	mockEndpoints()
}
