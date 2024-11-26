package auth

import (
	"context"
	"encoding/json"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/core/users"
	"github.com/brinestone/scholaris/core/webhooks"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/helpers"
)

var _ = pubsub.NewSubscription(webhooks.NewClerkUsers, "create-user-account", pubsub.SubscriptionConfig[dto.ClerkNewUserEventData]{
	Handler: onNewClerkUserCreated,
})

func onNewClerkUserCreated(ctx context.Context, data dto.ClerkNewUserEventData) (err error) {
	var dob *string
	rawJson, _ := json.Marshal(data)
	jsonData := string(rawJson)
	primaryEmailIndex, _ := helpers.Find(data.EmailAddresses, func(a dto.ClerkEmailAddress) bool {
		return a.ID == data.PrimaryEmailAddressID
	})

	var req = dto.NewExternalUserRequest{
		FirstName:  data.FirstName,
		LastName:   data.LastName,
		ExternalId: data.ID,
		Provider:   "clerk",
		Emails: helpers.Map[dto.ClerkEmailAddress, dto.ExternalUserEmailAddressData](data.EmailAddresses, func(a dto.ClerkEmailAddress) dto.ExternalUserEmailAddressData {
			return dto.ExternalUserEmailAddressData{
				Email:      a.EmailAddress,
				Verified:   a.Verification.Status == "verified",
				ExternalId: &a.ID,
				Primary:    a.ID == data.PrimaryEmailAddressID,
			}
		}),
		Phones: helpers.Map[dto.ClerkPhoneNumber, dto.ExternalUserPhoneData](data.PhoneNumbers, func(a dto.ClerkPhoneNumber) dto.ExternalUserPhoneData {
			return dto.ExternalUserPhoneData{
				Phone:      a.PhoneNumber,
				Verified:   a.Verification.Status == "verfified",
				ExternalId: &a.Id,
				Primary:    a.Id == *data.PrimaryPhoneNumberID,
			}
		}),
		Gender:       &data.Gender,
		Dob:          dob,
		Avatar:       &data.ImageURL,
		ProviderData: &jsonData,
	}
	var res *dto.NewUserResponse
	res, err = users.NewExternalUser(ctx, req)
	if err != nil {
		return
	}

	SignUps.Publish(ctx, UserSignedUp{
		Email:  data.EmailAddresses[primaryEmailIndex].EmailAddress,
		UserId: res.UserId,
	})
	return
}
