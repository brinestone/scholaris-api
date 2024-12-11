package permissions

import (
	"time"

	"encore.dev/beta/auth"
	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/util"
)

var relationsCache = cache.NewListKeyspace[string, string](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "relations/:key",
	DefaultExpiry: cache.ExpireIn(time.Hour * 2),
})

func relationsCacheKey(uid auth.UID, relations ...string) (ans string) {
	args := append(relations, string(uid))
	ans = util.HashThese(args...)
	return
}
