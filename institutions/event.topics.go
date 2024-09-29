package institutions

import "encore.dev/pubsub"

type InstitutionCreated struct{}

var NewInstitutions = pubsub.NewTopic[*InstitutionCreated]("new-institution", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
