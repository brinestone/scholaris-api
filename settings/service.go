// CRUD endpoints for user-defined and system generated settings.
package settings

import (
	"context"

	"github.com/brinestone/scholaris/dto"
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
	return nil, nil
}
