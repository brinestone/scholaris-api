package notifications

import (
	"context"
	"time"

	"encore.dev/pubsub"
	"github.com/brinestone/scholaris/core/urlshortener"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/tenants"
)

var _ = pubsub.NewSubscription(tenants.TenantInvites, "send-tenant-invite-emails", pubsub.SubscriptionConfig[*tenants.MemberInvited]{
	Handler: onNewMemberInvited,
})

func onNewMemberInvited(ctx context.Context, msg *tenants.MemberInvited) (err error) {
	err = SendEmail(ctx, dto.SendEmailRequest{
		To:     msg.Email,
		ToName: msg.DisplayName,
		Data: map[string]string{
			"inviteeName":        msg.DisplayName,
			"tenantName":         msg.TenantName,
			"tenantUrl":          "",
			"inviteUrl":          msg.Url,
			"invitationDeadline": msg.Deadline.Format(time.DateOnly),
		},
		TemplateId: "d-90749aa869484cbc9b5e8c29a47dfb0a",
	})
	if err != nil {
		return
	}

	maxClicks := 1
	window := time.Until(msg.Deadline)
	urlshortener.ShortenUrl(ctx, dto.ShortenUrlRequest{
		Url:       msg.Url,
		ErrorUrl:  &msg.ErrorUrl,
		MaxClicks: &maxClicks,
		Window:    &window,
	})
	return
}
