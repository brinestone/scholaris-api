package institutions

import (
	"context"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/settings"
)

var _ = pubsub.NewSubscription(NewInstitutions, "create-default-institution-settings", pubsub.SubscriptionConfig[*InstitutionCreated]{
	Handler: createDefaultSettingsOnNewInstitution,
})

var _ = pubsub.NewSubscription(settings.UpdatedSettings, "update-acacdemic-year-creation-cron", pubsub.SubscriptionConfig[settings.SettingUpdatedEvent]{
	Handler: assertAutoAcademicYearCreationCronJobs,
})

func assertAutoAcademicYearCreationCronJobs(ctx context.Context, msg settings.SettingUpdatedEvent) error {
	res, err := settings.GetSettingsInternal(ctx, dto.GetSettingsInternalRequest{
		Owner:     msg.Owner,
		OwnerType: msg.OwnerType,
		Ids:       msg.Ids,
	})
	if err != nil {
		return err
	}

	for _, v := range res.Settings {
		switch v.Key {
		case dto.SKAcademicYearAutoCreation:

		}
	}
	return nil
}

func createDefaultSettingsOnNewInstitution(ctx context.Context, msg *InstitutionCreated) error {
	return defineInstitutionDefaultSettings(ctx, msg.Id)
}
