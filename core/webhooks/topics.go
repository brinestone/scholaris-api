package webhooks

import (
	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
)

var NewClerkUsers = pubsub.NewTopic[dto.ClerkNewUserEventData]("new-clerk-user", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
