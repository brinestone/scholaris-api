package users

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/models"
)

var cluster = cache.NewCluster("scholaris", cache.ClusterConfig{
	EvictionPolicy: cache.AllKeysLRU,
})

var idCache = cache.NewStructKeyspace[uint64, models.User](cluster, cache.KeyspaceConfig{
	KeyPattern:    "read/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

var emailCache = cache.NewStructKeyspace[string, models.User](cluster, cache.KeyspaceConfig{
	KeyPattern:    "read-:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})
