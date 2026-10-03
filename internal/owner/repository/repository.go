package owner_repository

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// password_hash is never selected here: the owner domain only deals with the profile.
const ownerColumns = `id, email, nickname, COALESCE(gender, ''), COALESCE(avatar_url, ''), is_profile_public,
	is_gender_public, is_avatar_public, created_at`

type OwnerRepository struct {
	pool *pgxpool.Pool
}

func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

func (r *OwnerRepository) GetOwner(ctx context.Context, id string) (core_domain.Owner, error) {
	owner, err := scanOwner(r.pool.QueryRow(ctx, `SELECT `+ownerColumns+` FROM owners WHERE id = $1`, id))
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("select owner %s: %w", id, core_postgres.MapError(err))
	}
	return owner, nil
}

// UpdateProfile saves the editable profile fields; empty gender/avatar_url are stored as NULL.
func (r *OwnerRepository) UpdateProfile(ctx context.Context, owner core_domain.Owner) (core_domain.Owner, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE owners
		SET nickname = $2,
		    gender = NULLIF($3, ''),
		    avatar_url = NULLIF($4, ''),
		    is_gender_public = $5,
		    is_avatar_public = $6
		WHERE id = $1
		RETURNING `+ownerColumns,
		owner.Id, owner.Nickname, owner.Gender, owner.AvatarUrl, owner.Visibility.Gender, owner.Visibility.AvatarUrl,
	)
	updated, err := scanOwner(row)
	if err != nil {
		return core_domain.Owner{}, fmt.Errorf("update owner %s: %w", owner.Id, core_postgres.MapError(err))
	}
	return updated, nil
}

func scanOwner(row pgx.Row) (core_domain.Owner, error) {
	var owner core_domain.Owner
	err := row.Scan(&owner.Id, &owner.Email, &owner.Nickname, &owner.Gender, &owner.AvatarUrl, &owner.IsProfilePublic,
		&owner.Visibility.Gender, &owner.Visibility.AvatarUrl, &owner.CreatedAt)
	return owner, err
}
