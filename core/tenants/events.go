package tenants

import (
	"time"

	"encore.dev/pubsub"
)

type TenantCreated struct {
	Date time.Time
}

var NewTenantsTopic = pubsub.NewTopic[*TenantCreated]("create-tenant", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.ExactlyOnce,
})
