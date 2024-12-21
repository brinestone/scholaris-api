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
	maxClicks := 1
	window := time.Until(msg.Deadline)
	res, err := urlshortener.ShortenUrl(ctx, dto.ShortenUrlRequest{
		Url:       msg.Url,
		MaxClicks: &maxClicks,
		Window:    &window,
	})
	if err != nil {
		return
	}

	data := map[string]string{
		"inviteeName":        msg.DisplayName,
		"tenantName":         msg.TenantName,
		"inviteUrl":          res.ShortenedUrl,
		"invitationDeadline": msg.Deadline.Format(time.DateOnly),
	}
	err = SendEmail(ctx, dto.SendEmailRequest{
		To:         msg.Email,
		ToName:     msg.DisplayName,
		Data:       data,
		TemplateId: "d-90749aa869484cbc9b5e8c29a47dfb0a",
	})
	return
}
