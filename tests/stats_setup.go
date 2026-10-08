package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pkgclickhouse "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/clickhouse"
)

// OpenStatsConn returns a ClickHouse connection to the test container and
// guarantees the stats schema exists.
//
// Dua hal yang perlu diperhatikan:
//
//  1. Skema di-apply lewat pkgclickhouse.ApplySchema — DDL yang sama persis
//     dengan yang dipakai stats-reader/stats-writer saat startup (go:embed dari
//     pkg/clickhouse/schema.sql). Jadi test tidak bisa menyimpang dari
//     produksi, dan tabel tidak perlu ditulis ulang secara manual di sini.
//  2. Modul testcontainers ClickHouse hanya membuat *database*-nya
//     (WithDatabase), tidak pernah membuat tabel. Tanpa langkah ini seluruh
//     suite stats gagal dengan "Table ... doesn't exist".
//
// Koneksi dibuat langsung dari kredensial container, bukan lewat
// pkgclickhouse.NewClient, supaya test tidak perlu menyetel viper global.
func (ts *TestSuite) OpenStatsConn() (clickhouse.Conn, error) {
	addr, err := ts.CHContainer.ConnectionHost(ts.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve clickhouse endpoint: %w", err)
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: ts.CHContainer.DbName,
			Username: ts.CHContainer.User,
			Password: ts.CHContainer.Password,
		},
		DialTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(ts.Ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping clickhouse at %s: %w", addr, err)
	}

	if err := pkgclickhouse.ApplySchema(ts.Ctx, conn, ts.Logger); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to apply clickhouse schema: %w", err)
	}

	return conn, nil
}

// StatsTables adalah ketujuh tabel stats yang dibuat pkg/clickhouse/schema.sql.
// Dipakai untuk truncate antar-suite dan untuk memastikan tiap tipe event
// benar-benar mendarat di tabelnya sendiri.
//
// Kalau menambah tabel di schema.sql, tambahkan juga di sini — tidak ada yang
// menghubungkan kedua daftar itu secara otomatis.
var StatsTables = []string{
	"card_events",
	"merchant_events",
	"saldo_events",
	"topup_events",
	"transaction_events",
	"transfer_events",
	"withdraw_events",
}

// TruncateStatsTables mengosongkan seluruh tabel stats supaya tiap suite mulai
// dari state yang bersih. `IF EXISTS` membuatnya aman walau dipanggil sebelum
// ApplySchema.
func TruncateStatsTables(ctx context.Context, conn clickhouse.Conn) error {
	for _, table := range StatsTables {
		if err := conn.Exec(ctx, "TRUNCATE TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	return nil
}
