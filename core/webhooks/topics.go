package webhooks

import (
	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
)

var NewClerkUsers = pubsub.NewTopic[dto.ClerkEvent[dto.ClerkNewUserEventData]]("new-clerk-user-event", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
