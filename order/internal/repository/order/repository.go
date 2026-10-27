package order

import (
	"github.com/jackc/pgx/v5/pgxpool"

	def "inventory-go/order/internal/repository"
)

var _ def.OrderRepository = (*repository)(nil)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}
