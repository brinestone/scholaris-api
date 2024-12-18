// Endpoints for user Notifications
package notifications

import (
	"context"

	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
)

var secrets struct {
	SendgridKey string
}

type notifier interface {
	Notify(models.Notification) error
}

//encore:service
type Service struct {
	notifier notifier
}

//lint:ignore U1000 Intentionally ignored linting message
func initService() (s *Service, err error) {
	s = &Service{
		notifier: NewSendGridNotifier("no-reply@scholaris.space", "Scholaris Team", secrets.SendgridKey),
	}
	return
}

//encore:api private method=POST path=/notifications/email
func (s *Service) SendEmail(ctx context.Context, req dto.SendEmailRequest) error {
	notification := models.Notification{
		Subject: req.Subject,
		RecepientInfo: map[string]string{
			"name":    req.ToName,
			"address": req.To,
		},
		Data: req.Data,
	}
	if req.IsContentHtml {
		notification.Meta["htmlContent"] = req.Body
	} else {
		notification.Content = req.Body
	}

	if len(req.TemplateId) > 0 {
		notification.Meta["templateId"] = req.TemplateId
	}

	s.notifier.Notify(notification)
	return nil
}
