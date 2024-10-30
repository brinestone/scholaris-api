package billing

import "context"

type VerifyTransactionResponse struct {
	IsCleared     bool
	TransactionId uint64
}

type VerifyTransactionRequest struct {
	VerificationToken string `query:"token"`
}

type MakePaymentResponse struct {
	VerificationToken string
}

// Verifies whether a transaction is cleared or not
//
//encore:api private path=/billing/check method=GET
func VerifyTransaction(ctx context.Context, req VerifyTransactionRequest) (*VerifyTransactionResponse, error) {
	return &VerifyTransactionResponse{
		IsCleared:     true,
		TransactionId: 1,
	}, nil
}

// Get transaction token
//
//encore:api auth path=/billing/pay method=POST
func MakePayment(ctx context.Context) (*MakePaymentResponse, error) {
	return &MakePaymentResponse{
		VerificationToken: "valid",
	}, nil
}
