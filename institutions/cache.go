package institutions

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/noop"
	"github.com/brinestone/scholaris/models"
)

// var cluster = cache.NewCluster("scholaris", cache.ClusterConfig{
// 	EvictionPolicy: cache.AllKeysLRU,
// })

var institutionCache = cache.NewStructKeyspace[uint64, models.Institution](noop.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "institutions/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})
