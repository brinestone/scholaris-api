package notifications

import (
	"context"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/tenants"
)

var _ = pubsub.NewSubscription(tenants.TenantInvites, "send-tenant-invite-emails", pubsub.SubscriptionConfig[*tenants.MemberInvited]{
	Handler: onNewMemberInvited,
})

func onNewMemberInvited(ctx context.Context, msg *tenants.MemberInvited) (err error) {
	SendEmail(ctx, dto.SendEmailRequest{
		To:     msg.Email,
		ToName: msg.DisplayName,
		Data: map[string]string{
			"inviteeName":        msg.DisplayName,
			"tenantName":         "Foo Academy",
			"tenantUrl":          "https://example.com",
			"inviteUrl":          "https://example.com",
			"invitationDeadline": "12/02/2024",
		},
	})
	return
}
