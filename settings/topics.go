package settings

import (
	"time"

	"encore.dev/pubsub"
)

type SettingUpdatedEvent struct {
	Owner     uint64
	Ids       []uint64
	OwnerType string
	Timestamp time.Time
}

var UpdatedSettings = pubsub.NewTopic[SettingUpdatedEvent]("setting-updated", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
