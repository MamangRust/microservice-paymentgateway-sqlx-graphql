package main

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/apps"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "withdraw-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Port:           50060,
		DBCluster:      database.FinancialCluster,
		RedisCluster:   "REDIS_3",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
