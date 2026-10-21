package model

import "errors"

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrOrderPaid      = errors.New("order is already paid and cannot be cancelled")
	ErrEmptyPartUuids = errors.New("part_uuids must not be empty")
	ErrPartsNotFound  = errors.New("some parts were not found in the inventory")
)
