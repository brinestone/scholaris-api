package notifications

import (
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
	message := mail.NewSingleEmail(from, n.Subject, to, n.Content, n.Meta["htmlContent"])
	message.TemplateID = n.Meta["templateId"]

	_, err = s.client.Send(message)
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
