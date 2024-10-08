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

var NewInstitutions = pubsub.NewTopic[*InstitutionCreated]("new-institution", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
