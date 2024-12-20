package urlshortener_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/brinestone/scholaris/core/urlshortener"
	"github.com/brinestone/scholaris/dto"
	"github.com/stretchr/testify/assert"
)

func TestShortenUrl(t *testing.T) {
	url := gofakeit.URL()
	cnt := gofakeit.IntRange(10, 100)
	var maxCount *int
	var errUrl *string
	var window *time.Duration

	if gofakeit.Bool() {
		maxCount = &cnt
	}
	if gofakeit.Bool() {
		tmp := gofakeit.URL()
		errUrl = &tmp
	}
	if gofakeit.Bool() {
		tmp := time.Hour * time.Duration(gofakeit.Float32Range(10.0, 50))
		window = &tmp
	}

	res, err := urlshortener.ShortenUrl(context.TODO(), dto.ShortenUrlRequest{
		Url:       url,
		MaxClicks: maxCount,
		ErrorUrl:  errUrl,
		Window:    window,
	})

	if assert.Nil(t, err) {
		assert.Equal(t, url, res.OriginalUrl)
		assert.NotEmpty(t, res.ShortenedUrl)
	}
}
