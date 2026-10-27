package order

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
	repoModel "inventory-go/order/internal/repository/model"
)

func (r *repository) Get(ctx context.Context, uuid string) (model.Order, error) {
	query, args, err := sq.Select(columnUuid, columnUserUuid, columnPartUuids, columnTotalPrice,
		columnTransactionUuid, columnPaymentMethod, columnStatus).
		From(tableName).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{columnUuid: uuid}).
		Limit(1).
		ToSql()
	if err != nil {
		return model.Order{}, err
	}

	var o repoModel.Order
	err = r.pool.QueryRow(ctx, query, args...).Scan(&o.Uuid, &o.UserUuid, &o.PartUuids, &o.TotalPrice,
		&o.TransactionUuid, &o.PaymentMethod, &o.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, model.ErrOrderNotFound
		}

		return model.Order{}, err
	}

	return repoConverter.OrderToModel(o), nil
}
