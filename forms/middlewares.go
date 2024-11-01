package forms

import (
	"encore.dev/middleware"
	"github.com/brinestone/scholaris/core/auth"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

//encore:middleware target=tag:needs_captcha_ver
func VerifyCaptcha(req middleware.Request, next middleware.Next) middleware.Response {
	p, ok := req.Data().Payload.(models.CaptchaVerifiable)
	if !ok {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	if err := auth.VerifyCaptchaToken(req.Context(), auth.VerifyCaptchaRequest{
		Token: p.GetCaptchaToken(),
	}); err != nil {
		return middleware.Response{
			Err: &util.ErrCaptchaError,
		}
	}

	return next(req)
}
