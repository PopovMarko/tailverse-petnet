package pet_repository

import (
	"context"
	"fmt"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const petColumns = `id, owner_id, name, COALESCE(breed, ''), species, birth_date, COALESCE(approx_address, ''), created_at`

type PetRepository struct {
	pool *pgxpool.Pool
}

func NewPetRepository(pool *pgxpool.Pool) *PetRepository {
	return &PetRepository{pool: pool}
}

func (r *PetRepository) CreatePet(ctx context.Context, pet core_domain.Pet) (core_domain.Pet, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO pets (owner_id, name, species, breed, birth_date, approx_address)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''))
		RETURNING `+petColumns,
		pet.OwnerId, pet.Name, pet.Species, pet.Breed, nullableDate(pet.BirthDate), pet.ApproxAddress,
	)
	created, err := scanPet(row)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("insert pet: %w", core_postgres.MapError(err))
	}
	return created, nil
}

func (r *PetRepository) GetPet(ctx context.Context, id string) (core_domain.Pet, error) {
	pet, err := scanPet(r.pool.QueryRow(ctx, `SELECT `+petColumns+` FROM pets WHERE id = $1`, id))
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("select pet %s: %w", id, core_postgres.MapError(err))
	}
	return pet, nil
}

func (r *PetRepository) GetPetsByOwner(ctx context.Context, ownerId string) ([]core_domain.Pet, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+petColumns+` FROM pets WHERE owner_id = $1 ORDER BY created_at, id`, ownerId)
	if err != nil {
		return nil, fmt.Errorf("select pets of owner %s: %w", ownerId, err)
	}
	pets, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.Pet, error) {
		return scanPet(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan pets: %w", err)
	}
	return pets, nil
}

func (r *PetRepository) UpdatePet(ctx context.Context, pet core_domain.Pet) (core_domain.Pet, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pets
		SET name = $2, species = $3, breed = NULLIF($4, ''), birth_date = $5, approx_address = NULLIF($6, '')
		WHERE id = $1
		RETURNING `+petColumns,
		pet.Id, pet.Name, pet.Species, pet.Breed, nullableDate(pet.BirthDate), pet.ApproxAddress,
	)
	updated, err := scanPet(row)
	if err != nil {
		return core_domain.Pet{}, fmt.Errorf("update pet %s: %w", pet.Id, core_postgres.MapError(err))
	}
	return updated, nil
}

func (r *PetRepository) DeletePet(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM pets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete pet %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete pet %s: %w", id, core_postgres.MapError(pgx.ErrNoRows))
	}
	return nil
}

func scanPet(row pgx.Row) (core_domain.Pet, error) {
	var (
		pet       core_domain.Pet
		birthDate *time.Time
	)
	if err := row.Scan(&pet.Id, &pet.OwnerId, &pet.Name, &pet.Breed, &pet.Species, &birthDate, &pet.ApproxAddress, &pet.CreatedAt); err != nil {
		return core_domain.Pet{}, err
	}
	if birthDate != nil {
		pet.BirthDate = *birthDate
	}
	return pet, nil
}

func nullableDate(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// GetPetCards returns pet name + owner nickname for every existing pet id. Nickname is empty when the owner hides the profile.
func (r *PetRepository) GetPetCards(ctx context.Context, petIds []string) (map[string]core_domain.PetCard, error) {
	cards := make(map[string]core_domain.PetCard, len(petIds))
	if len(petIds) == 0 {
		return cards, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, CASE WHEN o.is_profile_public THEN o.nickname ELSE '' END
		FROM pets p
		JOIN owners o ON o.id = p.owner_id
		WHERE p.id = ANY($1::uuid[])`,
		petIds,
	)
	if err != nil {
		return nil, fmt.Errorf("select pet cards: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var card core_domain.PetCard
		if err := rows.Scan(&card.PetId, &card.PetName, &card.OwnerNickname); err != nil {
			return nil, fmt.Errorf("scan pet card: %w", err)
		}
		cards[card.PetId] = card
	}
	return cards, rows.Err()
}
