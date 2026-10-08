package database

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewClientWithPrefix opens a *sqlx.DB for the bounded context identified by
// prefix (one of the *Cluster constants, e.g. "DB_IDENTITY"). The prefix
// resolves its own <Prefix>_HOST / _PORT / _NAME; only the username and password
// fall back to the generic DB_* keys. A missing host, port or database name is
// an error so a misconfigured service never silently connects to another
// context's database.
func NewClientWithPrefix(logger logger.LoggerInterface, prefix string) (*sqlx.DB, error) {
	if prefix == "" {
		prefix = "DB"
	}

	dbDriver := firstNonEmpty(
		viper.GetString(fmt.Sprintf("%s_DRIVER", prefix)),
		viper.GetString("DB_DRIVER"),
	)
	if dbDriver != "postgres" && dbDriver != "pgx" {
		logger.Error("sqlx only supports PostgreSQL", zap.String("DB_DRIVER", dbDriver))
		return nil, fmt.Errorf("sqlx only supports PostgreSQL, got: %s", dbDriver)
	}

	host := viper.GetString(fmt.Sprintf("%s_HOST", prefix))
	port := viper.GetString(fmt.Sprintf("%s_PORT", prefix))
	dbname := viper.GetString(fmt.Sprintf("%s_NAME", prefix))
	user := firstNonEmpty(
		viper.GetString(fmt.Sprintf("%s_USERNAME", prefix)),
		viper.GetString("DB_USERNAME"),
	)
	password := firstNonEmpty(
		viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix)),
		viper.GetString("DB_PASSWORD"),
	)

	if host == "" || port == "" || dbname == "" {
		logger.Error("Incomplete database configuration",
			zap.String("cluster", prefix),
			zap.String("host", host),
			zap.String("port", port),
			zap.String("dbname", dbname),
		)
		return nil, fmt.Errorf("incomplete database configuration for %s (host=%q port=%q dbname=%q)",
			prefix, host, port, dbname)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password,
	)

	maxOpenConns := viper.GetInt(fmt.Sprintf("%s_MAX_OPEN_CONNS", prefix))
	if maxOpenConns <= 0 {
		maxOpenConns = viper.GetInt("DB_MAX_OPEN_CONNS")
		if maxOpenConns <= 0 {
			maxOpenConns = 100
		}
	}

	minIdleConns := viper.GetInt(fmt.Sprintf("%s_MIN_IDLE_CONNS", prefix))
	if minIdleConns <= 0 {
		minIdleConns = viper.GetInt("DB_MIN_IDLE_CONNS")
		if minIdleConns <= 0 {
			minIdleConns = 50
		}
	}
	if minIdleConns > maxOpenConns {
		minIdleConns = maxOpenConns
	}

	connMaxLifetime := viper.GetDuration(fmt.Sprintf("%s_CONN_MAX_LIFETIME", prefix))
	if connMaxLifetime == 0 {
		connMaxLifetime = viper.GetDuration("DB_CONN_MAX_LIFETIME")
		if connMaxLifetime == 0 {
			connMaxLifetime = time.Hour
		}
	}

	connMaxIdleTime := viper.GetDuration(fmt.Sprintf("%s_CONN_MAX_IDLE_TIME", prefix))
	if connMaxIdleTime == 0 {
		connMaxIdleTime = viper.GetDuration("DB_CONN_MAX_IDLE_TIME")
		if connMaxIdleTime == 0 {
			connMaxIdleTime = 30 * time.Minute
		}
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err), zap.String("prefix", prefix))
		return nil, fmt.Errorf("failed to connect to database for %s: %w", prefix, err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(minIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// DB_*_HEALTH_CHECK_PERIOD is no longer read: database/sql has no
	// pgxpool-style health-check loop equivalent.

	logger.Debug("Database connection pool established successfully",
		zap.String("prefix", prefix),
		zap.String("DB_DRIVER", "pgx"),
		zap.Int("MaxOpenConns", maxOpenConns),
		zap.Int("MaxIdleConns", minIdleConns),
		zap.Duration("ConnMaxLifetime", connMaxLifetime),
		zap.Duration("ConnMaxIdleTime", connMaxIdleTime),
	)

	return db, nil
}

// firstNonEmpty returns the first non-empty string, or "" if all are empty.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
