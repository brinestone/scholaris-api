package auth

import (
	"encore.dev/middleware"
	"encore.dev/rlog"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Captcha verification middleware
//
//encore:middleware target=tag:needs_captcha_ver
func ValidateCaptcha(request middleware.Request, next middleware.Next) (res middleware.Response) {
	res = next(request)

	data, ok := request.Data().Payload.(models.CaptchaVerifiable)
	if !ok {
		res = middleware.Response{
			Err: &util.ErrCaptchaError,
		}
		return
	}

	if err := verifyCaptcha(data.GetCaptchaToken()); err != nil {
		rlog.Error("captcha error", "msg", err.Error())
		res = middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	return
}
