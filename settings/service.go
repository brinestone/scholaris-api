// CRUD endpoints for user-defined and system generated settings.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"encore.dev/beta/auth"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Updates a setting
//
//encore:api auth method=POST path=/settings/:owner tag:can_update_settings
func UpdateSetting(ctx context.Context, owner uint64, req dto.UpdateSettingsRequest) error {
	return nil
}

// Gets an owner's settings
//
//encore:api auth method=GET path=/settings/:owner tag:can_view_settings
func GetSettings(ctx context.Context, owner uint64) (*dto.GetSettingsResponse, error) {
	var settings *dto.GetSettingsResponse
	var err error
	uid, _ := auth.UserID()
	s, err := settingsCache.Get(ctx, cacheKey(uid, owner))
	if errors.Is(err, cache.Miss) {
		perms, err := permissions.ListRelations(ctx, dto.ListRelationsRequest{
			Actor:    dto.IdentifierString(dto.PTUser, uid),
			Relation: models.PermCanView,
			Type:     string(dto.PTSetting),
		})
		if err != nil {
			rlog.Error(util.MsgCallError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}

		ids := perms.Relations[dto.PTSetting]
		if len(ids) == 0 {
			return &dto.GetSettingsResponse{}, nil
		}

		mods, err := findSettingsFromDb(ctx, owner, ids...)
		if err != nil {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}

		dtos := settingsToDto(mods...)
		settings = &dto.GetSettingsResponse{
			Settings: dtos,
		}

		if err := settingsCache.Set(ctx, cacheKey(uid, owner), *settings); err != nil {
			rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
		}
	} else if err != nil {
		rlog.Error(util.MsgCacheAccessError, "msg", err.Error())
		return nil, &util.ErrUnknown
	} else {
		settings = &s
	}
	return settings, nil
}

func findSettingsFromDb(ctx context.Context, owner uint64, ids ...uint64) ([]*models.Setting, error) {
	args := make([]any, len(ids)+1)
	args[0] = owner
	placeholders := make([]string, len(ids))

	for i, v := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i+1] = v
	}

	query := fmt.Sprintf(`
		SELECT
			s.id,
			s.label,
			s.description,
			s.key,
			s.multi_values,
			s.created_at,
			s.updated_at,
			s.parent,
			s.owner,
			s.owner_type,
			s.created_by,
			s.overridable,
			COALESCE(json_agg(json_build_object(
				'id', so.id,
				'label', so.label,
				'value', so.label,
				'setting', so.setting
			)) FILTER (WHERE so.setting IS NOT NULL), '[]') as options,
			COALESCE(json_agg(json_build_object(
				'id', sv.id,
				'setting', sv.setting,
				'value', sv.value,
				'setAt', sv.set_at,
				'setBy', sv.set_by
			)) FILTER (WHERE sv.setting IS NOT NULL), '[]') as values
		FROM
			settings s
		LEFT JOIN
			setting_options so
				ON so.setting = s.id
		LEFT JOIN
			setting_values sv
				ON sv.setting = s.id
		WHERE
			s.owner = $1 AND s.system_generated=false AND id IN (%s)
		GROUP BY
			s.id
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.Setting
	for rows.Next() {
		setting := new(models.Setting)
		var optionsJson, valuesJson string
		if err := rows.Scan(&setting.Id, &setting.Label, &setting.Description, &setting.Key, &setting.MultiValues, &setting.CreatedAt, &setting.UpdatedAt, &setting.Parent, &setting.Owner, &setting.OwnerType, &setting.CreatedBy, &setting.Overridable, &optionsJson, &valuesJson); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(valuesJson), &setting.Values); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(optionsJson), &setting.Options); err != nil {
			return nil, err
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

func settingsToDto(s ...*models.Setting) []dto.Setting {
	var ans = make([]dto.Setting, len(s))

	for i, v := range s {
		var setting dto.Setting
		setting.Id = v.Id
		setting.Label = v.Label
		setting.Key = v.Key
		setting.MultiValues = v.MultiValues
		setting.CreatedAt = v.CreatedAt
		setting.UpdatedAt = v.UpdatedAt
		setting.Owner = v.Owner
		setting.OwnerType = v.OwnerType
		setting.Overrideable = v.Overridable
		setting.CreatedBy = v.CreatedBy
		setting.Options = make([]dto.SettingOption, len(v.Options))
		setting.Values = make([]dto.SettingValue, len(v.Values))
		if v.Description.Valid {
			setting.Description = &v.Description.String
		}
		if v.Parent.Valid {
			tmp := uint64(v.Parent.Int64)
			setting.Parent = &tmp
		}

		for j, w := range v.Options {
			var option dto.SettingOption
			option.Id = w.Id
			option.Label = w.Label
			option.Setting = w.Setting
			option.Value = w.Value

			setting.Options[j] = option
		}

		for j, w := range v.Values {
			var value dto.SettingValue
			value.Id = w.Id
			value.SetAt = w.SetAt
			value.SetBy = w.SetBy
			value.Setting = w.Setting

			setting.Values[j] = value
		}

		ans[i] = setting
	}

	return ans
}
