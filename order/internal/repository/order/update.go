package order

import (
	"context"

	sq "github.com/Masterminds/squirrel"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
)

func (r *repository) Update(ctx context.Context, order model.Order) error {
	o := repoConverter.OrderToRepoModel(order)

	query, args, err := sq.Update(tableName).
		PlaceholderFormat(sq.Dollar).
		Set(columnUserUuid, o.UserUuid).
		Set(columnPartUuids, o.PartUuids).
		Set(columnTotalPrice, o.TotalPrice).
		Set(columnTransactionUuid, o.TransactionUuid).
		Set(columnPaymentMethod, o.PaymentMethod).
		Set(columnStatus, o.Status).
		Set(columnUpdatedAt, sq.Expr("now()")).
		Where(sq.Eq{columnUuid: o.Uuid}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return model.ErrOrderNotFound
	}

	return nil
}
