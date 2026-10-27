-- +goose Up
create table orders (
    uuid             uuid primary key,
    user_uuid        uuid             not null,
    part_uuids       uuid[]           not null,
    total_price      double precision not null,
    transaction_uuid uuid,
    payment_method   text,
    status           text             not null,
    created_at       timestamptz      not null default now(),
    updated_at       timestamptz
);

create index orders_user_uuid_idx on orders (user_uuid);

-- +goose Down
drop table orders;
