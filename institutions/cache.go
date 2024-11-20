package institutions

import (
	"time"

	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/dto"
)

var institutionCache = cache.NewStructKeyspace[string, dto.Institution](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "institutions/:key",
	DefaultExpiry: cache.ExpireIn(5 * time.Minute),
})

//// var enrollmentCache = cache.NewStructKeyspace[uint64, dto.EnrollmentState](pkg.CacheCluster, cache.KeyspaceConfig{
//// 	KeyPattern:    "enrollments/:key",
//// 	DefaultExpiry: cache.ExpireIn(20 * time.Minute),
//// })
