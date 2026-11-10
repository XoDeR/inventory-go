package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"inventory-go/auth/internal/model"
)

const selectUser = `select uuid, login, email, password_hash, created_at, updated_at from users`

func (r *repository) GetByUUID(ctx context.Context, uuid string) (model.User, error) {
	return r.get(ctx, selectUser+` where uuid = $1`, uuid)
}

func (r *repository) GetByLogin(ctx context.Context, login string) (model.User, error) {
	return r.get(ctx, selectUser+` where login = $1`, login)
}

func (r *repository) get(ctx context.Context, query string, arg any) (model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&u.UUID, &u.Login, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}

		return model.User{}, fmt.Errorf("select user: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`select provider_name, target from user_notification_methods where user_uuid = $1 order by id`,
		u.UUID,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("select notification methods: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m model.NotificationMethod
		if err = rows.Scan(&m.ProviderName, &m.Target); err != nil {
			return model.User{}, fmt.Errorf("scan notification method: %w", err)
		}

		u.NotificationMethods = append(u.NotificationMethods, m)
	}

	if err = rows.Err(); err != nil {
		return model.User{}, fmt.Errorf("read notification methods: %w", err)
	}

	return u, nil
}
