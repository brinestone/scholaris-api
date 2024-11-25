package auth

import (
	"context"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/core/webhooks"
	"github.com/brinestone/scholaris/dto"
)

var _ = pubsub.NewSubscription(webhooks.ClerkEvents, "update-users", pubsub.SubscriptionConfig[dto.ClerkEvent]{
	Handler: onClerkEvent,
})

func onClerkEvent(ctx context.Context, msg dto.ClerkEvent) (err error) {

	return
}
