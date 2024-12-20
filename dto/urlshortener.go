package dto

import "time"

type ShortenUrlResponse struct {
	ShortenedUrl string `json:"shortened"`
	OriginalUrl  string `json:"originalUrl"`
}

type ShortenUrlRequest struct {
	Url       string         `json:"url"`
	MaxClicks *int           `json:"maxClicks,omitempty" encore:"optional"`
	ErrorUrl  *string        `json:"errorUrl,omitempty" encore:"optional"`
	Window    *time.Duration `json:"window,omitempty" encore:"optional"`
}
