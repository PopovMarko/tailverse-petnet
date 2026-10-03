package core_postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "github.com/PopovMarko/tailverse-petnet/internal/core/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

func NewPool(ctx context.Context, config PostgresConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	// Return every timestamptz in UTC, so API responses do not depend on the server's time zone.
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
		})
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres %s:%d: %w", config.Host, config.Port, err)
	}
	return pool, nil
}

// MapError converts pgx errors into core_errors so that the transport layer can pick a status code.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", err, core_errors.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return fmt.Errorf("%s: %w", pgErr.ConstraintName, core_errors.ErrConflict)
		case pgForeignKeyViolation:
			return fmt.Errorf("referenced entity does not exist (%s): %w", pgErr.ConstraintName, core_errors.ErrInvalidArgument)
		case pgCheckViolation:
			return fmt.Errorf("constraint %s violated: %w", pgErr.ConstraintName, core_errors.ErrInvalidArgument)
		}
	}
	return err
}
