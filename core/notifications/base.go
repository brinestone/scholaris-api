package notifications

import (
	"context"
	"fmt"

	"encore.dev/pubsub"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/core/tenants"
)

var _ = pubsub.NewSubscription(tenants.NewTenants, "email-sending", pubsub.SubscriptionConfig[*tenants.TenantCreated]{
	Handler: sendEmailNotification,
})

var _ = pubsub.NewSubscription(auth.SignUps, "onboarding-email", pubsub.SubscriptionConfig[*auth.UserSignedUp]{
	Handler: sendOnboardingEmail,
})

func sendOnboardingEmail(ctx context.Context, msg *auth.UserSignedUp) error {
	return nil
}

func sendEmailNotification(ctx context.Context, msg *tenants.TenantCreated) error {
	rlog.Debug(fmt.Sprint(msg.Id))
	return nil
}
