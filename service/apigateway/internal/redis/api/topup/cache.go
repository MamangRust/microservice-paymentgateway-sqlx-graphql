package topup_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type TopupMencach interface {
	TopupQueryCache
	TopupCommandCache
}

type mencache struct {
	TopupQueryCache
	TopupCommandCache
}

func NewTopupMencache(cacheStore *cache.CacheStore) TopupMencach {
	return &mencache{
		TopupQueryCache:   NewTopupQueryCache(cacheStore),
		TopupCommandCache: NewTopupCommandCache(cacheStore),
	}
}
