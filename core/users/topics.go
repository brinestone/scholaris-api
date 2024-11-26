package users

import (
	"time"

	"encore.dev/pubsub"
)

type UserDeleted struct {
	UserId    uint64
	Timestamp time.Time
}

type UserAccountCreated struct {
	UserId    uint64
	AccountId uint64
	Timestamp time.Time
	NewUser   bool
}

var DeletedUsers = pubsub.NewTopic[UserDeleted]("user-deleted", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var NewUsers = pubsub.NewTopic[UserAccountCreated]("account-added", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
