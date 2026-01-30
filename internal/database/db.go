package database

import (
	"context"
	"database/sql"

	"github.com/Secure-Website-Builder/Backend/internal/models"
)

type DB struct {
	db      *sql.DB
	Queries *models.Queries
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		db:      db,
		Queries: models.New(db),
	}
}

func (d *DB) QueryContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *DB) RunInTx(ctx context.Context, fn func(q *models.Queries) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := d.Queries.WithTx(tx)

	if err := fn(qtx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
