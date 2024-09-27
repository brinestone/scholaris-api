package notifications

import (
	"context"

	"encore.dev/pubsub"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/tenants"
)

var _ = pubsub.NewSubscription(tenants.NewTenantsTopic, "email-sending", pubsub.SubscriptionConfig[*tenants.TenantCreated]{
	Handler: sendEmailNotification,
})

var _ = pubsub.NewSubscription(tenants.NewTenantsTopic, "push-notification-sending", pubsub.SubscriptionConfig[*tenants.TenantCreated]{
	Handler: sendPushNotification,
})

func sendEmailNotification(ctx context.Context, msg *tenants.TenantCreated) error {
	rlog.Debug(msg.Date.String())
	return nil
}

func sendPushNotification(ctx context.Context, msg *tenants.TenantCreated) error {
	rlog.Debug(msg.Date.String())
	return nil
}
