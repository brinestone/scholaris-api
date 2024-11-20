package tenants

import (
	"context"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/core/users"
)

var _ = pubsub.NewSubscription(users.DeletedUsers, "purge-owned-tenants", pubsub.SubscriptionConfig[users.UserDeleted]{
	Handler: onUserAccountDeleted,
})

func onUserAccountDeleted(ctx context.Context, msg users.UserDeleted) (err error) {
	// TODO: Remove tenants owned by this user.
	// TODO: Remove the tenants owned by this users if they're the only member.
	return
}
