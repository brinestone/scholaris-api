package users

import (
	"time"

	"encore.dev/pubsub"
)

type UserDeleted struct {
	UserId    uint64
	Timestamp time.Time
}

var DeletedUsers = pubsub.NewTopic[UserDeleted]("user-deleted", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
