package models

import "io"

// TODO: Use this later
type NotificationAttachment struct {
	Thumbnail *string
	Name      string
	io.ReadCloser
}

type Notification struct {
	Subject             string
	Content             string
	Meta, RecepientInfo map[string]string
	Data                map[string]any
}
