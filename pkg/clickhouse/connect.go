package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// resolveAddr merakit alamat ClickHouse dari CLICKHOUSE_ADDR atau
// CLICKHOUSE_HOST + CLICKHOUSE_PORT.
func resolveAddr() string {
	addr := viper.GetString("CLICKHOUSE_ADDR")
	if addr == "" {
		host := viper.GetString("CLICKHOUSE_HOST")
		if host == "" {
			host = "clickhouse"
		}
		port := viper.GetString("CLICKHOUSE_PORT")
		if port == "" {
			port = "9000"
		}
		addr = fmt.Sprintf("%s:%s", host, port)
	}
	return addr
}

// EnsureDatabase membuat CLICKHOUSE_DATABASE kalau belum ada. NewClient memakai
// database itu sebagai default, jadi database harus sudah ada sebelum client
// bisa ping — karena itu fungsi ini dipanggil lebih dulu. No-op untuk database
// "default" atau kalau CLICKHOUSE_DATABASE kosong.
func EnsureDatabase(l logger.LoggerInterface) error {
	dbName := viper.GetString("CLICKHOUSE_DATABASE")
	if dbName == "" || dbName == "default" {
		return nil
	}

	// Koneksi bootstrap sengaja TANPA Auth.Database: kalau database-nya belum
	// ada, mengirimnya sebagai default justru membuat handshake gagal.
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{resolveAddr()},
		Auth: clickhouse.Auth{
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		l.Error("Failed to open ClickHouse connection for database bootstrap", zap.Error(err))
		return fmt.Errorf("failed to open clickhouse for database bootstrap: %w", err)
	}
	defer conn.Close()

	if err := conn.Exec(context.Background(), "CREATE DATABASE IF NOT EXISTS "+dbName); err != nil {
		l.Error("Failed to create ClickHouse database", zap.Error(err), zap.String("database", dbName))
		return fmt.Errorf("failed to create clickhouse database %s: %w", dbName, err)
	}

	l.Debug("ClickHouse database ensured", zap.String("database", dbName))
	return nil
}

func NewClient(l logger.LoggerInterface) (clickhouse.Conn, error) {
	addr := resolveAddr()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: viper.GetString("CLICKHOUSE_DATABASE"),
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		l.Error("Failed to open ClickHouse connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			l.Error("ClickHouse exception during ping", 
				zap.Uint32("code", uint32(exception.Code)), 
				zap.String("message", exception.Message),
			)
		}
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	l.Debug("ClickHouse connection established successfully", zap.String("addr", addr))
	return conn, nil
}
