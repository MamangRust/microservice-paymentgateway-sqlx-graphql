// Package database constants for the three bounded contexts of the payment
// gateway. Each context runs on its own PostgreSQL instance; the database keeps
// the context name so the same identifier works in env keys, migrations and
// dashboards. A service selects its own through the cluster prefix (see the
// *Cluster constants) so that one table has exactly one owner.
package database

const (
	// IdentityDB hosts users, roles, user_roles, refresh_tokens and reset_tokens.
	IdentityDB = "pg_identity"
	// PaymentDB hosts merchants, merchant_documents, cards, the card_* tables,
	// saldos, balance_ledger and reconciliation_queue.
	PaymentDB = "pg_payment"
	// FinancialDB hosts transactions, topups, transfers, withdraws,
	// idempotency_records, operation state and outbox_events.
	FinancialDB = "pg_financial"
)

// Cluster* are the env prefixes passed to NewClientWithPrefix. Each prefix
// resolves its own <Prefix>_HOST, _PORT and _NAME — one PostgreSQL instance per
// bounded context, fronted by its own PgBouncer. Only the username and password
// fall back to the generic DB_* keys; DB_HOST / DB_PORT / DB_NAME are
// deliberately absent from the environment so that a missing per-context key
// fails fast instead of silently targeting another bounded context's database.
const (
	IdentityCluster  = "DB_IDENTITY"
	PaymentCluster   = "DB_PAYMENT"
	FinancialCluster = "DB_FINANCIAL"
)
