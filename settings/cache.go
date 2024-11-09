package settings

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"encore.dev/beta/auth"
	"encore.dev/storage/cache"
	"github.com/brinestone/scholaris/core/pkg"
	"github.com/brinestone/scholaris/dto"
)

var settingsCache = cache.NewStructKeyspace[string, dto.GetSettingsResponse](pkg.CacheCluster, cache.KeyspaceConfig{
	KeyPattern: "settings/:key",
})

func cacheKey(user auth.UID, owner uint64, ownerType string) string {
	arg := md5.Sum([]byte(fmt.Sprintf("%v", []any{user, owner, ownerType})))
	return hex.EncodeToString(arg[:])
}
