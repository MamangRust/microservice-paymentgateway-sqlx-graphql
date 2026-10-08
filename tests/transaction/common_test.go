package transaction_test

import (
	"context"

	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
)

type mockAISecurityServer struct {
	pbAISecurity.UnimplementedAISecurityServiceServer
}

func (m *mockAISecurityServer) DetectFraud(ctx context.Context, in *pbAISecurity.FraudRequest) (*pbAISecurity.FraudResponse, error) {
	return &pbAISecurity.FraudResponse{
		TransactionId: in.TransactionId,
		RiskScore:     0.1,
		IsFraudulent:  false,
		Reason:        "Mocked safe transaction",
	}, nil
}
