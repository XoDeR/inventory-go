package main

import (
	"context"
	paymentV1 "inventory-go/shared/pkg/proto/payment/v1"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serverAddress = "localhost:50052"

func main() {
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Printf("failed to close connection: %v\n", cerr)
		}
	}()

	client := paymentV1.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.PayOrder(ctx, &paymentV1.PayOrderRequest{
		OrderUuid:     uuid.NewString(),
		UserUuid:      uuid.NewString(),
		PaymentMethod: paymentV1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	if err != nil {
		log.Printf("failed to pay order: %v\n", err)
		return
	}

	log.Printf("PayOrder: transaction_uuid: %s\n", resp.GetTransactionUuid())
}
