package forms

import (
	"time"

	"encore.dev/pubsub"
)

type FormCreated struct {
	Id        uint64
	Timestamp time.Time
}

type FormPublished struct {
	Id        uint64
	Timestamp time.Time
}

var NewForms = pubsub.NewTopic[FormCreated]("new-form", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var PublishedForms = pubsub.NewTopic[FormPublished]("form-published", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
