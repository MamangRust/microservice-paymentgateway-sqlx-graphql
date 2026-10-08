package transaction_cache

import "time"

const (
	transactionAllCacheKey          = "transaction:all:page:%d:pageSize:%d:search:%s"
	transactionByIdCacheKey         = "transaction:id:%d"
	transactionActiveCacheKey       = "transaction:active:page:%d:pageSize:%d:search:%s"
	transactionTrashedCacheKey      = "transaction:trashed:page:%d:pageSize:%d:search:%s"
	transactionByMerchantIdCacheKey = "transaction:merchant_id:%d"

	ttlDefault = 5 * time.Minute
)
