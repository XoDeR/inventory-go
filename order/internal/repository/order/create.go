package order

import (
	"context"

	sq "github.com/Masterminds/squirrel"

	"inventory-go/order/internal/model"
	repoConverter "inventory-go/order/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, order model.Order) error {
	o := repoConverter.OrderToRepoModel(order)

	query, args, err := sq.Insert(tableName).
		PlaceholderFormat(sq.Dollar).
		Columns(columnUuid, columnUserUuid, columnPartUuids, columnTotalPrice,
			columnTransactionUuid, columnPaymentMethod, columnStatus).
		Values(o.Uuid, o.UserUuid, o.PartUuids, o.TotalPrice,
			o.TransactionUuid, o.PaymentMethod, o.Status).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)

	return err
}
