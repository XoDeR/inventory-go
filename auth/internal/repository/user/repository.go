package user

import (
	"github.com/jackc/pgx/v5/pgxpool"

	def "inventory-go/auth/internal/repository"
)

var _ def.UserRepository = (*repository)(nil)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}
