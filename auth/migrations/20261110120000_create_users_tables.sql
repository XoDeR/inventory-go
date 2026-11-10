-- +goose Up
create table users (
    uuid          uuid primary key,
    login         text        not null unique,
    email         text        not null unique,
    password_hash text        not null,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz
);

create table user_notification_methods (
    id            bigserial primary key,
    user_uuid     uuid not null references users (uuid) on delete cascade,
    provider_name text not null,
    target        text not null
);

create index user_notification_methods_user_uuid_idx on user_notification_methods (user_uuid);

-- +goose Down
drop table user_notification_methods;
drop table users;
