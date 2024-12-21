package notifications

import (
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridNotifier struct {
	from       string
	senderName string
	client     *sendgrid.Client
}

func (s *SendGridNotifier) Notify(n models.Notification) (err error) {
	from := mail.NewEmail(s.senderName, s.from)
	to := mail.NewEmail(n.RecepientInfo["name"], n.RecepientInfo["address"])
	message := mail.NewSingleEmail(from, n.Subject, to, "", "")
	message.TemplateID = n.Meta["templateId"]
	message.Personalizations = []*mail.Personalization{
		{
			To:                  []*mail.Email{to},
			From:                from,
			DynamicTemplateData: n.Data,
			Subject:             n.Subject,
		},
	}
	res, err := s.client.Send(message)
	rlog.Debug("sendgrid", "message", message, "result", res)
	return
}

func NewSendGridNotifier(senderEmail, senderName, apiKey string) (ans *SendGridNotifier) {
	ans = &SendGridNotifier{
		from:       senderEmail,
		senderName: senderName,
		client:     sendgrid.NewSendClient(apiKey),
	}
	return
}
