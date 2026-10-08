package clickhouse

import (
	"context"
	_ "embed"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"go.uber.org/zap"
)

// schemaSQL dibaca langsung dari schema.sql (bukan disalin jadi const) supaya
// DDL yang dijalankan runtime tidak pernah menyimpang dari file yang di-mount
// sebagai initdb di Compose/Kubernetes.
//
//go:embed schema.sql
var schemaSQL string

// ApplySchema memastikan ketujuh tabel stats ada sebelum stats-writer menulis
// atau stats-reader membaca. Semua statement adalah CREATE TABLE IF NOT EXISTS,
// jadi aman dipanggil berulang kali (idempotent) — ini yang membuat tabel tetap
// ada walaupun volume ClickHouse sudah berisi data tapi tabelnya hilang,
// kondisi yang tidak bisa ditangani mekanisme initdb image ClickHouse.
func ApplySchema(ctx context.Context, conn clickhouse.Conn, l logger.LoggerInterface) error {
	for _, stmt := range strings.Split(schemaSQL, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := conn.Exec(ctx, stmt); err != nil {
			l.Error("Failed to apply ClickHouse schema statement", zap.Error(err), zap.String("statement", stmt))
			return err
		}
	}
	l.Debug("ClickHouse schema applied")
	return nil
}
