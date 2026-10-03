package services_repository

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxNearbyServices = 200

const serviceColumns = `id, provider_owner_id, title, description, category,
	ST_Y(location::geometry), ST_X(location::geometry), COALESCE(price, 0)::float8, created_at`

type ServiceOfferRepository struct {
	pool *pgxpool.Pool
}

func NewServiceOfferRepository(pool *pgxpool.Pool) *ServiceOfferRepository {
	return &ServiceOfferRepository{pool: pool}
}

func (r *ServiceOfferRepository) CreateServiceOffer(ctx context.Context, offer core_domain.ServiceOffer) (core_domain.ServiceOffer, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO services (provider_owner_id, title, description, category, location, price)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, $7)
		RETURNING `+serviceColumns,
		offer.ProviderOwnerId, offer.Title, offer.Description, offer.Category, offer.Location.Lng, offer.Location.Lat, offer.Price,
	)
	created, err := scanServiceOffer(row)
	if err != nil {
		return core_domain.ServiceOffer{}, fmt.Errorf("insert service: %w", core_postgres.MapError(err))
	}
	return created, nil
}

func (r *ServiceOfferRepository) GetServiceOffer(ctx context.Context, id string) (core_domain.ServiceOffer, error) {
	offer, err := scanServiceOffer(r.pool.QueryRow(ctx, `SELECT `+serviceColumns+` FROM services WHERE id = $1`, id))
	if err != nil {
		return core_domain.ServiceOffer{}, fmt.Errorf("select service %s: %w", id, core_postgres.MapError(err))
	}
	return offer, nil
}

func (r *ServiceOfferRepository) ListNearby(ctx context.Context, filter core_domain.ServiceOfferFilter) ([]core_domain.ServiceOffer, error) {
	rows, err := r.pool.Query(ctx, `
		WITH center AS (SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography AS point)
		SELECT `+serviceColumns+`
		FROM services, center
		WHERE ST_DWithin(location, center.point, $3)
		  AND ($4::text IS NULL OR category = $4)
		ORDER BY location <-> center.point
		LIMIT $5`,
		filter.Nearby.Center.Lng, filter.Nearby.Center.Lat, filter.Nearby.RadiusM, filter.Category, maxNearbyServices,
	)
	if err != nil {
		return nil, fmt.Errorf("select nearby services: %w", err)
	}
	offers, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.ServiceOffer, error) {
		return scanServiceOffer(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan services: %w", err)
	}
	return offers, nil
}

func scanServiceOffer(row pgx.Row) (core_domain.ServiceOffer, error) {
	var o core_domain.ServiceOffer
	err := row.Scan(&o.Id, &o.ProviderOwnerId, &o.Title, &o.Description, &o.Category, &o.Location.Lat, &o.Location.Lng, &o.Price, &o.CreatedAt)
	return o, err
}
