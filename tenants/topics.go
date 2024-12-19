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

type MemberInvited struct {
	TenantName  string
	Id          uint64
	Email       string
	DisplayName string
	Url         string
}

var NewTenants = pubsub.NewTopic[*TenantCreated]("new-tenant", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var DeletedTenants = pubsub.NewTopic[*TenantDeleted]("tenant-deleted", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var TenantInvites = pubsub.NewTopic[*MemberInvited]("tenant-member-invited", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
