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
