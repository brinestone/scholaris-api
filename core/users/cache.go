package users

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/models"
)

// var cluster = cache.NewCluster("scholaris", cache.ClusterConfig{
// 	EvictionPolicy: cache.AllKeysLRU,
// })

var idCache = cache.NewStructKeyspace[uint64, models.User](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "users/:key",
	DefaultExpiry: cache.ExpireIn(1 * time.Hour),
})

var emailCache = cache.NewStructKeyspace[string, models.User](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "users-:key",
	DefaultExpiry: cache.ExpireIn(1 * time.Hour),
})
