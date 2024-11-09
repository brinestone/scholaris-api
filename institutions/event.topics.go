package institutions

import (
	"time"

	"encore.dev/beta/auth"
	"encore.dev/pubsub"
)

type InstitutionCreated struct {
	Id        uint64
	CreatedBy auth.UID
	Timestamp time.Time
}

type EnrollmentPublished struct {
	Id          uint64
	Institution uint64
	Owner       auth.UID
}

var NewInstitutions = pubsub.NewTopic[*InstitutionCreated]("institution-created", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var PublishedEnrollments = pubsub.NewTopic[*EnrollmentPublished]("enrollment-published", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
