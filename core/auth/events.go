package auth

import "encore.dev/pubsub"

type UserSignedUp struct {
	Email  string
	UserId uint64
}

var SignUps = pubsub.NewTopic[*UserSignedUp]("sign-up", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
