package models

import "sync"

type OrderStorage struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string]*Order),
	}
}

func (s *OrderStorage) CreateOrder(order Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.Uuid] = &order
}

func (s *OrderStorage) UpdateOrder(order Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.Uuid] = &order
}

func (s *OrderStorage) GetOrder(uuid string) *Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[uuid]
	if !ok {
		return nil
	}
	return order
}
