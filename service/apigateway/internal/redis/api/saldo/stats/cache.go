package saldo_stats_cache

import "github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"

type SaldoStatsCache interface {
	SaldoStatsBalanceCache
	SaldoStatsTotalCache
}

type saldoStatsCache struct {
	SaldoStatsBalanceCache
	SaldoStatsTotalCache
}

func NewSaldoStatsCache(store *cache.CacheStore) SaldoStatsCache {
	return &saldoStatsCache{
		SaldoStatsBalanceCache: NewSaldoStatsBalanceCache(store),
		SaldoStatsTotalCache:   NewSaldoStatsTotalCache(store),
	}
}
