package models

import "time"

type Order struct {
	Uuid            string    `json:"uuid"`
	UserUuid        string    `json:"user_uuid"`
	PartUuids       []string  `json:"part_uuids"`
	TotalPrice      int       `json:"total_price"`
	TransactionUuid *string   `json:"transaction_uuid"`
	PaymentMethod   *string   `json:"payment_method"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
