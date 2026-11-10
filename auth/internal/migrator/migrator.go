package migrator

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// Up applies all pending migrations found in the root of migrationsFS.
func (m *Migrator) Up(ctx context.Context, migrationsFS fs.FS) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, m.db, migrationsFS)
	if err != nil {
		return err
	}

	_, err = provider.Up(ctx)

	return err
}
