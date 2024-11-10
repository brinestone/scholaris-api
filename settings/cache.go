package settings

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/dto"
)

var settingsCache = cache.NewStructKeyspace[string, dto.GetSettingsResponse](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern:    "settings/:key",
	DefaultExpiry: cache.ExpireIn(time.Second * 30),
})

func cacheKey(owner uint64, ownerType string) string {
	uid, _ := auth.UserID()
	arg := md5.Sum([]byte(fmt.Sprintf("%v", []any{uid, owner, ownerType})))
	return hex.EncodeToString(arg[:])
}
