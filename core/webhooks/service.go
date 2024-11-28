package webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"encore.dev/rlog"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/util"
	svix "github.com/svix/svix-webhooks/go"
)

var secrets struct {
	SvixSecret string
}

// Clerk Webhook
//
//encore:api public raw path=/webhooks/clerk
func ClerkWebhook(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
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

	if err = verifySvixWebhookRequest(req, requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		rlog.Warn(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", "invalid attempt", "data", event)
		return
	}

	dataAsJson, _ := json.Marshal(event.Data)
	switch event.Type {
	case dto.CEUserCreated:
		var eventData dto.ClerkNewUserEventData
		json.Unmarshal(dataAsJson, &eventData)
		_, err = NewClerkUsers.Publish(req.Context(), eventData)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
		json, _ := json.Marshal(&util.ErrUnknown)
		if _, err = w.Write(json); err != nil {
			rlog.Error(util.MsgWebhookError, "webhook", "ClerkWebhook", "msg", err.Error())
			return
		}
	}
}

func verifySvixWebhookRequest(r *http.Request, dataJson []byte) (err error) {
	headers := http.Header{}
	headers.Set("svix-id", r.Header.Get("svix-id"))
	headers.Set("svix-timestamp", r.Header.Get("svix-timestamp"))
	headers.Set("svix-signature", r.Header.Get("svix-signature"))

	wh, err := svix.NewWebhook(secrets.SvixSecret)
	if err != nil {
		return
	}

	err = wh.Verify(dataJson, headers)
	return
}
