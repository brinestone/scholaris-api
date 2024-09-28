package auth

import "encore.dev/pubsub"

type UserSignedUp struct {
	Email  string
	UserId int64
}

var SignUps = pubsub.NewTopic[*UserSignedUp]("user-signups", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})