// CRUD endpoints for user-defined and system generated settings.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/cache"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/core/permissions"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
	"github.com/lib/pq"
)

// Internally fetches settings
//
//encore:api private method=GET path=/settings/internal
func GetSettingsInternal(ctx context.Context, req dto.GetSettingsInternalRequest) (*dto.GetSettingsResponse, error) {
	mods, err := findSettingsFromDb(ctx, req.Owner, req.OwnerType, true, req.Ids...)
	if err != nil {
		return nil, err
	}

	dtos := settingsToDto(mods...)
	return &dto.GetSettingsResponse{
		Settings: dtos,
	}, nil
}

// Sets the value(s) of a setting.  Intended for internal APIs
//
//encore:api auth method=PUT path=/settings/set tag:can_set_setting
func SetSettingValues(ctx context.Context, req dto.SetSettingValueRequest) error {
	userId, _ := auth.UserID()
	uid, _ := strconv.ParseUint(string(userId), 10, 64)
	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		return &util.ErrUnknown
	}

	if err := updateSettingValues(ctx, tx, req.Owner, uid, req.OwnerType, req.Updates...); err != nil {
		tx.Rollback()
		if errs.Convert(err) == nil {
			return err
		} else {
			rlog.Error(util.MsgDbAccessError, "err", err)
		}
		return &util.ErrUnknown
	}

	defer tx.Commit()
	return nil
}

// Updates settings. Intended for internal APIs
//
//encore:api private method=POST path=/settings/internal
func UpdateSettingsInternal(ctx context.Context, req dto.UpdateSettingsRequest) error {
	var uid uint64 = 0
	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	ids, err := updateSettings(ctx, tx, req.Owner, uid, req.OwnerType, req.Updates...)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	var updates []dto.PermissionUpdate
	for _, id := range ids {
		updates = append(updates,
			dto.PermissionUpdate{
				Actor:    dto.IdentifierString(dto.PermissionType(req.GetOwnerType()), req.Owner),
				Relation: models.PermOwner,
				Target:   dto.IdentifierString(dto.PTSetting, id),
			}, dto.PermissionUpdate{
				Actor:    dto.IdentifierString(dto.PTUser, uid),
				Relation: models.PermEditor,
				Target:   dto.IdentifierString(dto.PTSetting, id),
			})
	}

	if err := permissions.SetPermissions(ctx, dto.UpdatePermissionsRequest{
		Updates: updates,
	}); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		return &util.ErrUnknown
	} else if err := tx.Commit(); err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	if len(ids) > 0 {
		UpdatedSettings.Publish(ctx, SettingUpdatedEvent{
			Owner:     req.Owner,
			Ids:       ids,
			OwnerType: req.OwnerType,
			Timestamp: time.Now(),
		})
	}
	return nil
}

// Updates settings (public API)
//
//encore:api auth method=POST path=/settings tag:can_update_settings tag:needs_captcha_ver
func UpdateSettings(ctx context.Context, req dto.UpdateSettingsRequest) error {
	uid, _ := auth.UserID()
	user, _ := strconv.ParseUint(string(uid), 10, 64)
	tx, err := db.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	ids, err := updateSettings(ctx, tx, req.Owner, user, req.OwnerType, req.Updates...)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}
	if err := tx.Commit(); err != nil {
		rlog.Error(util.MsgDbAccessError, "msg", err.Error())
		return &util.ErrUnknown
	}

	if len(ids) > 0 {
		UpdatedSettings.Publish(ctx, SettingUpdatedEvent{
			Owner:     req.Owner,
			Ids:       ids,
			OwnerType: req.OwnerType,
			Timestamp: time.Now(),
		})
	}
	return nil
}

// Gets an owner's settings
//
//encore:api auth method=GET path=/settings tag:can_view_settings
func FindSettings(ctx context.Context, req dto.GetSettingsRequest) (*dto.GetSettingsResponse, error) {
	var settings *dto.GetSettingsResponse
	var err error
	uid, _ := auth.UserID()
	s, err := settingsCache.Get(ctx, cacheKey(req.Owner, req.OwnerType))
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
			return nil, &util.ErrNotFound
		}

		mods, err := findSettingsFromDb(ctx, req.Owner, req.OwnerType, false, ids...)
		if err != nil {
			rlog.Error(util.MsgDbAccessError, "msg", err.Error())
			return nil, &util.ErrUnknown
		}

		dtos := settingsToDto(mods...)
		settings = &dto.GetSettingsResponse{
			Settings: dtos,
		}

		if err := settingsCache.Set(ctx, cacheKey(req.Owner, req.OwnerType), *settings); err != nil {
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

func findSettingsFromDb(ctx context.Context, owner uint64, ownerType string, internal bool, ids ...uint64) ([]*models.Setting, error) {
	query := "SELECT * FROM func_get_owner_settings($1,$2,$3,$4,$4);"
	rows, err := db.Query(ctx, query, pq.Array(ids), owner, ownerType, internal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.Setting
	for rows.Next() {
		s := new(models.Setting)
		var optionsJson, valuesJson string
		if err := rows.Scan(&s.Id, &s.Description, &s.Key, &s.MultiValues, &s.CreatedAt, &s.UpdatedAt, &s.Parent, &s.Owner, &s.OwnerType, &s.CreatedBy, &s.Overridable, &s.SystemGenerated, &optionsJson, &valuesJson); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(valuesJson), &s.Values); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(optionsJson), &s.Options); err != nil {
			return nil, err
		}

		settings = append(settings, s)
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
			value.Index = w.Index
			setting.Values[j] = value
		}

		ans[i] = setting
	}

	return ans
}

func updateSettings(ctx context.Context, tx *sqldb.Tx, owner, user uint64, ownerType string, req ...dto.SettingUpdate) ([]uint64, error) {
	var ids []uint64
	for _, v := range req {
		query := `
			INSERT INTO
				settings(
					label,
					description,
					key,
					multi_values,
					parent,
					owner,
					owner_type,
					created_by,
					overridable
				)
			VALUES
				($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT 
				(owner,owner_type,key)
			DO
				UPDATE SET
					label=$1,
					description=$2,
					multi_values=$4,
					parent=$5,
					updated_at=DEFAULT,
					overridable=$9
			RETURNING id
			;
		`

		var id uint64
		if err := tx.QueryRow(ctx, query, v.Label, v.Description, v.Key, v.MultiValues, v.Parent, owner, ownerType, user, v.Overridable).Scan(&id); err != nil {
			return nil, err
		}

		if len(v.Options) > 0 {
			optionQuery := `
				INSERT INTO
					setting_options(label,value,setting)
				VALUES
					($1,$2,$3);
			`
			for _, w := range v.Options {
				if _, err := tx.Exec(ctx, optionQuery, w.Label, w.Value, id); err != nil {
					return nil, err
				}
			}
		}

		ids = append(ids, id)
	}
	return ids, nil
}

func updateSettingValues(ctx context.Context, tx *sqldb.Tx, owner, user uint64, ownerType string, req ...dto.SettingValueUpdate) error {
	for _, v := range req {
		q1 := `
			SELECT
				id
			FROM
				settings
			WHERE
				key=$1 AND owner=$2 AND owner_type=$3 AND system_generated=false;
		`
		var settingId uint64
		if err := db.QueryRow(ctx, q1, v.Key, owner, ownerType).Scan(&settingId); err != nil {
			if errors.Is(err, sqldb.ErrNoRows) {
				return &errs.Error{
					Code:    errs.FailedPrecondition,
					Message: fmt.Sprintf("unknown setting %s", v.Key),
				}
			}
		}

		query := `
			INSERT INTO
				setting_values(
					set_by,
					value,
					set_at,
					setting
					value_index
				)
			VALUES
				($1,$2,DEFAULT,$3,$4)
			ON CONFLICT 
				(setting,value_index)
			DO
				UPDATE SET
					value=$2,
					set_by=$1,
					set_at=DEFAULT;
		`
		for _, value := range v.Value {
			if _, err := tx.Exec(ctx, query, user, value, settingId, value.Index); err != nil {
				return err
			}
		}
	}
	return nil
}
