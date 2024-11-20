package forms

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/dto"
)

var formCache = cache.NewStructKeyspace[uint64, dto.FormConfig](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "forms/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

var formsCache = cache.NewStructKeyspace[string, dto.GetFormsResponse](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "form-requests/:key",
	DefaultExpiry: cache.ExpireIn(2 * time.Hour),
})

var questionsCache = cache.NewStructKeyspace[string, dto.GetFormQuestionsResponse](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "form-questions/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})
