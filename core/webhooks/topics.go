package webhooks

import (
	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
)

var ClerkEvents = pubsub.NewTopic[dto.ClerkEvent]("clerk-user-event", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
