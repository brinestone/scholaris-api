package notifications

import (
	"context"

	"github.com/brinestone/scholaris/dto"
)

// TODO use an emailing service for implementation.

//encore:api private method=POST path=/notifications/email
func SendEmail(ctx context.Context, req dto.SendEmailRequest) error {
	return nil
}
