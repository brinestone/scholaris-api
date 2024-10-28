package tenants

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/noop"
	"github.com/brinestone/scholaris/models"
)

var tenantCache = cache.NewStructKeyspace[uint64, models.Tenant](noop.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "tenants/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})
