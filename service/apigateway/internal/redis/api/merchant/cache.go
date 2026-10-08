package merchant_cache

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
)

type mencache struct {
	MerchantQueryCache
	MerchantCommandCache
}

type MerchantMencache interface {
	MerchantQueryCache
	MerchantCommandCache
}

func NewMerchantMencache(cacheStore *cache.CacheStore) MerchantMencache {
	return &mencache{
		MerchantQueryCache:   NewMerchantQueryCache(cacheStore),
		MerchantCommandCache: NewMerchantCommandCache(cacheStore),
	}
}
