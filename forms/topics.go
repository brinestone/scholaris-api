package forms

import (
	"time"

	"encore.dev/pubsub"
)

type FormEvent struct {
	Id        uint64
	Timestamp time.Time
	Owner     uint64
	OwnerType string
}
type FormPublished struct {
	Id        uint64
	Owner     uint64
	OwnerType string
	Timestamp time.Time
}
type FormDeleted struct {
	Id        uint64
	Owner     uint64
	OwnerType string
	Timestamp time.Time
}
type ResponseSubmitted struct {
	Form      uint64
	Response  uint64
	Timestamp time.Time
	Owner     uint64
	OwnerType string
}

var NewForms = pubsub.NewTopic[FormEvent]("form-created", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var PublishedForms = pubsub.NewTopic[FormPublished]("form-published", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var DeletedForms = pubsub.NewTopic[FormDeleted]("form-deleted", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var FormSubmissions = pubsub.NewTopic[ResponseSubmitted]("response-submitted", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
