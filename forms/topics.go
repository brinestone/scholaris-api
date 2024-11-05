package forms

import (
	"time"

	"encore.dev/pubsub"
)

type FormCreated struct {
	Id        uint64
	Timestamp time.Time
}

var NewForms = pubsub.NewTopic[FormCreated]("new-form", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
