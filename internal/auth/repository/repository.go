package auth_repository

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ownerColumns = `id, email, password_hash, nickname, COALESCE(gender, ''), COALESCE(avatar_url, ''), is_profile_public,
	is_gender_public, is_avatar_public, created_at`

type OwnerRepository struct {
	pool *pgxpool.Pool
}

func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

func (r *OwnerRepository) CreateOwner(ctx context.Context, owner core_domain.Owner) (core_domain.Owner, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO owners (email, password_hash, nickname, gender, avatar_url, is_gender_public, is_avatar_public)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7)
		RETURNING `+ownerColumns,
		owner.Email, owner.PasswordHash, owner.Nickname, owner.Gender, owner.AvatarUrl,
		owner.Visibility.Gender, owner.Visibility.AvatarUrl,
	)
	created, err := scanOwner(row)
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("insert owner: %w", core_postgres.MapError(err))
	}
	return created, nil
}

func (r *OwnerRepository) GetOwnerByEmail(ctx context.Context, email string) (core_domain.Owner, error) {
	owner, err := scanOwner(r.pool.QueryRow(ctx, `SELECT `+ownerColumns+` FROM owners WHERE lower(email) = lower($1)`, email))
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("select owner by email: %w", core_postgres.MapError(err))
	}
	return owner, nil
}

func (r *OwnerRepository) GetOwner(ctx context.Context, id string) (core_domain.Owner, error) {
	owner, err := scanOwner(r.pool.QueryRow(ctx, `SELECT `+ownerColumns+` FROM owners WHERE id = $1`, id))
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("select owner %s: %w", id, core_postgres.MapError(err))
	}
	return owner, nil
}

func scanOwner(row pgx.Row) (core_domain.Owner, error) {
	var owner core_domain.Owner
	err := row.Scan(&owner.Id, &owner.Email, &owner.PasswordHash, &owner.Nickname, &owner.Gender, &owner.AvatarUrl, &owner.IsProfilePublic,
		&owner.Visibility.Gender, &owner.Visibility.AvatarUrl, &owner.CreatedAt)
	return owner, err
}
