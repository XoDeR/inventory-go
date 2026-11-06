// Package events holds the contracts of the events exchanged between services through Kafka.
// The events are JSON-encoded, see the contracts/ directory for the description.
package events

import (
	"encoding/json"
	"fmt"
)

const (
	// TopicOrderPaid is the default name of the topic with OrderPaid events.
	TopicOrderPaid = "order.paid"
	// TopicComputerAssembled is the default name of the topic with ComputerAssembled events.
	TopicComputerAssembled = "assembly.computer-assembled"
)

// OrderPaid is published by OrderService after an order was successfully paid.
type OrderPaid struct {
	EventUUID       string `json:"event_uuid"`
	OrderUUID       string `json:"order_uuid"`
	UserUUID        string `json:"user_uuid"`
	PaymentMethod   string `json:"payment_method"`
	TransactionUUID string `json:"transaction_uuid"`
}

// ComputerAssembled is published by AssemblyService after the computer of an order was assembled.
type ComputerAssembled struct {
	EventUUID    string `json:"event_uuid"`
	OrderUUID    string `json:"order_uuid"`
	UserUUID     string `json:"user_uuid"`
	BuildTimeSec int64  `json:"build_time_sec"`
}

// Encode serializes an event to JSON.
func Encode(event any) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to encode event: %w", err)
	}

	return data, nil
}

// Decode parses a JSON-encoded event into dst.
func Decode(data []byte, dst any) error {
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("failed to decode event: %w", err)
	}

	return nil
}
