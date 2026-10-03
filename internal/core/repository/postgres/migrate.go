package core_postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationLockId is an arbitrary constant for pg_advisory_lock, so that two API instances never migrate at once.
const migrationLockId = 7_318_205

// Migrate applies every "*.up.sql" file from migrations that is not yet recorded in schema_migrations.
// Files are applied in lexical order, each one in its own transaction.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) ([]string, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockId); err != nil {
		return nil, fmt.Errorf("take migration lock: %w", err)
	}
	defer conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLockId)

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := fs.Glob(migrations, "*.up.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files)

	var applied []string
	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")

		var exists bool
		if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists); err != nil {
			return applied, fmt.Errorf("check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		sql, err := fs.ReadFile(migrations, file)
		if err != nil {
			return applied, fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return applied, fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return applied, fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback(ctx)
			return applied, fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, fmt.Errorf("commit migration %s: %w", version, err)
		}
		applied = append(applied, version)
	}
	return applied, nil
}
