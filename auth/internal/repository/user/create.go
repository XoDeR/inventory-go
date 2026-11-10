package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"inventory-go/auth/internal/model"
)

const uniqueViolation = "23505"

func (r *repository) Create(ctx context.Context, user model.User) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx,
		`insert into users (uuid, login, email, password_hash) values ($1, $2, $3, $4)`,
		user.UUID, user.Login, user.Email, user.PasswordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return model.ErrUserAlreadyExists
		}

		return fmt.Errorf("insert user: %w", err)
	}

	batch := &pgx.Batch{}
	for _, m := range user.NotificationMethods {
		batch.Queue(
			`insert into user_notification_methods (user_uuid, provider_name, target) values ($1, $2, $3)`,
			user.UUID, m.ProviderName, m.Target,
		)
	}

	br := tx.SendBatch(ctx, batch)
	for range user.NotificationMethods {
		if _, err = br.Exec(); err != nil {
			_ = br.Close()
			return fmt.Errorf("insert notification method: %w", err)
		}
	}

	if err = br.Close(); err != nil {
		return fmt.Errorf("close batch: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
