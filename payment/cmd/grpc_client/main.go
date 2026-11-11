package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
)

const serverAddress = "localhost:50052"

func main() {
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			slog.Error("failed to close connection", "error", cerr)
		}
	}()

	client := paymentV1.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// the service needs a session: log in via the auth service and pass SESSION_UUID=<session_uuid>
	if sessionUUID := os.Getenv("SESSION_UUID"); sessionUUID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "session-uuid", sessionUUID)
	}

	resp, err := client.PayOrder(ctx, &paymentV1.PayOrderRequest{
		OrderUuid:     uuid.NewString(),
		UserUuid:      uuid.NewString(),
		PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	if err != nil {
		slog.Error("failed to pay order", "error", err)
		return
	}

	slog.Info("PayOrder", "transaction_uuid", resp.GetTransactionUuid())
}
