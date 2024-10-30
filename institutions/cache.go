package institutions

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/dto"
)

// var cluster = cache.NewCluster("scholaris", cache.ClusterConfig{
// 	EvictionPolicy: cache.AllKeysLRU,
// })

var institutionCache = cache.NewStructKeyspace[string, dto.InstitutionDto](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "institutions/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

var enrollmentQuestionCache = cache.NewStructKeyspace[uint64, dto.EnrollmentQuestion](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "eqs/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})
