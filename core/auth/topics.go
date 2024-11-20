package auth

import (
	"time"

	"encore.dev/pubsub"
)

type UserSignedUp struct {
	Email  string
	UserId uint64
}

type UserSignedIn struct {
	Email     string
	UserId    uint64
	Timestamp time.Time
}

var SignUps = pubsub.NewTopic[UserSignedUp]("sign-up", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var SignIns = pubsub.NewTopic[UserSignedIn]("sign-in", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
