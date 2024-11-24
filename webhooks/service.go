package webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"encore.dev/rlog"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/util"
)

// Clerk Webhook
//
//encore:api public raw path=/auth/clerk/webhook
func ClerkWebhook(w http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
		json, _ := json.Marshal(&util.ErrUnknown)

		if _, err = w.Write(json); err != nil {
			rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
			return
		}
	}

	var event = new(dto.ClerkEvent)
	if err := json.Unmarshal(requestBody, event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
		json, _ := json.Marshal(&util.ErrUnknown)

		if _, err = w.Write(json); err != nil {
			rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
			return
		}
	}
	rlog.Debug("clerk event received", "event", event)
}
