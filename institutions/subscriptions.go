package institutions

import (
	"context"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/forms"
	"github.com/brinestone/scholaris/settings"
)

var _ = pubsub.NewSubscription(NewInstitutions, "create-default-institution-settings", pubsub.SubscriptionConfig[*InstitutionCreated]{
	Handler: createDefaultSettingsOnNewInstitution,
})

var _ = pubsub.NewSubscription(settings.UpdatedSettings, "update-acacdemic-year-creation-cron", pubsub.SubscriptionConfig[settings.SettingUpdatedEvent]{
	Handler: assertAutoAcademicYearCreationCronJobs,
})

var _ = pubsub.NewSubscription(forms.DeletedForms, "update-enrollment-forms", pubsub.SubscriptionConfig[forms.FormDeleted]{
	Handler: updateEnrollmentFormsOnFormDeleted,
})

func updateEnrollmentFormsOnFormDeleted(ctx context.Context, msg forms.FormDeleted) (err error) {
	if msg.OwnerType != string(dto.PTInstitution) {
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return
	}

	if err = deleteEnrollmentFormRegistrationByFormId(ctx, tx, msg.Id, msg.Owner); err != nil {
		tx.Rollback()
	}
	return
}

func assertAutoAcademicYearCreationCronJobs(ctx context.Context, msg settings.SettingUpdatedEvent) error {
	res, err := settings.FindSettingsInternal(ctx, dto.GetSettingsInternalRequest{
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
