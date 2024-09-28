package tenants

import (
	"time"

	"encore.dev/beta/auth"
	"encore.dev/pubsub"
)

type TenantCreated struct {
	CreatedBy *auth.UID
	Id        uint64
}

type TenantDeleted struct {
	Id        uint64
	DeletedAt time.Time
}

var NewTenants = pubsub.NewTopic[*TenantCreated]("new-tenant", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.ExactlyOnce,
})

var DeletedTenants = pubsub.NewTopic[*TenantDeleted]("delete-tenant", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.ExactlyOnce,
})
