package settings

import (
	"context"

	"encore.dev/pubsub"
)

var _ = pubsub.NewSubscription(UpdatedSettings, "cache-purge", pubsub.SubscriptionConfig[SettingUpdatedEvent]{
	Handler: func(ctx context.Context, msg SettingUpdatedEvent) error {
		settingsCache.Delete(ctx, cacheKey(msg.Owner, msg.OwnerType))
		return nil
	},
})
