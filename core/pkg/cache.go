package pkg

import "encore.dev/storage/cache"

var CacheCluster = cache.NewCluster("scholaris", cache.ClusterConfig{
	EvictionPolicy: cache.AllKeysLRU,
})
