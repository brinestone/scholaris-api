package urlshortener

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"math/rand"

	"encore.dev"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/brinestone/scholaris/dto"
	"github.com/brinestone/scholaris/models"
	"github.com/brinestone/scholaris/util"
)

// Shortened URLs
//
//encore:api public raw method=GET path=/short/:key
func AccessShortenedUrl(w http.ResponseWriter, req *http.Request) {
	key := encore.CurrentRequest().PathParams.Get("key")

	link, err := findLink(req.Context(), key)
	if errors.Is(err, sqldb.ErrNoRows) {
		errs.HTTPError(w, &util.ErrNotFound)
		return
	} else if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	if !link.IsUsable {
		if link.ErrorRedirect.Valid {
			http.Redirect(w, req, link.ErrorRedirect.String, http.StatusPermanentRedirect)
			return
		}
		errs.HTTPError(w, &util.ErrForbidden)
		return
	}

	tx, err := linksDb.Begin(req.Context())
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}

	if err = updateShortLinkClickCount(req.Context(), tx, key); err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		errs.HTTPError(w, &util.ErrUnknown)
		return
	}
	defer tx.Commit()

	http.Redirect(w, req, link.SuccessRedirect, http.StatusMovedPermanently)
}

// Shortens any URL (private api)
//
//encore:api private method=POST path=/short
func ShortenUrl(ctx context.Context, req dto.ShortenUrlRequest) (ans dto.ShortenUrlResponse, err error) {

	tx, err := linksDb.Begin(ctx)
	if err != nil {
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	key := generateShortKey()
	err = shortenUrl(ctx, tx, req.Url, key, req.MaxClicks, req.ErrorUrl, req.Window)
	if err != nil {
		tx.Rollback()
		rlog.Error(util.MsgDbAccessError, "err", err)
		err = &util.ErrUnknown
		return
	}

	tx.Commit()

	shortenedUrl, err := url.JoinPath(encore.Meta().APIBaseURL.String(), "short", key)
	if err != nil {
		rlog.Error("url error", "err", err)
		err = &util.ErrUnknown
		return
	}

	ans = dto.ShortenUrlResponse{
		ShortenedUrl: shortenedUrl,
		OriginalUrl:  req.Url,
	}
	return
}

func shortenUrl(ctx context.Context, tx *sqldb.Tx, url, key string, maxClicks *int, errorUrl *string, window *time.Duration) (err error) {
	query := `
		INSERT INTO 
			links("key",max_clicks,error_redirect,success_redirect,"window")
		VALUES($1,$2,$3,$4,$5);
	`
	var windowStr *string
	if window != nil {
		tmp := fmt.Sprintf("%f hours", window.Hours())
		windowStr = &tmp
	}
	_, err = tx.Exec(ctx, query, key, maxClicks, errorUrl, url, windowStr)
	return
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func scanShortLink(s util.RowScanner) (ans *models.ShortLink, err error) {
	ans = new(models.ShortLink)

	err = s.Scan(&ans.Key, &ans.ClickCount, &ans.MaxClicks, &ans.ErrorRedirect, &ans.SuccessRedirect, &ans.CreatedAt, &ans.UpdatedAt, &ans.ExpiresAt, &ans.IsUsable)
	if err != nil {
		err = errs.Wrap(err, "scan error")
	}
	return
}

func findLink(ctx context.Context, key string) (ans *models.ShortLink, err error) {
	query := `
		SELECT
			*
		FROM
			vw_AllLinks
		WHERE
			key=$1;
	`
	ans, err = scanShortLink(linksDb.QueryRow(ctx, query, key))
	return
}

func updateShortLinkClickCount(ctx context.Context, tx *sqldb.Tx, key string) (err error) {
	query := `
		UPDATE links
		SET
			click_count=click_count+1,
			updated_at=DEFAULT
		WHERE
			key=$1;
	`
	_, err = tx.Exec(ctx, query, key)
	return
}
