package adapter

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// GuardSetter is implemented by every gRPC adapter so WithDependencyGuard can
// attach a guard without changing each constructor's client parameter.
type GuardSetter interface {
	SetGuard(*resilience.DependencyGuard)
}

// GuardOption configures a GuardSetter (a gRPC adapter). Adapters accept zero
// or more options at construction; the guard wraps every outbound gRPC call
// with a per-call deadline, bulkhead and circuit breaker.
type GuardOption func(GuardSetter)

// WithDependencyGuard returns a GuardOption that attaches a dependency guard
// (per-call timeout + circuit breaker + bulkhead) to a gRPC adapter.
// Passing nil disables guarding.
//
// Usage:
//
//	guard := resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)
//	saldoAdapter := adapter.NewSaldoAdapter(q, c, adapter.WithDependencyGuard(guard))
func WithDependencyGuard(guard *resilience.DependencyGuard) GuardOption {
	return func(s GuardSetter) {
		s.SetGuard(guard)
	}
}
