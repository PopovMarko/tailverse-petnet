package walkspot_repository

import (
	"context"
	"fmt"
	"math"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// maxNearbySpots caps the map response; the closest spots come first.
const maxNearbySpots = 200

const walkSpotColumns = `id, name, ST_Y(location::geometry), ST_X(location::geometry), tags, created_at`

type WalkSpotRepository struct {
	pool *pgxpool.Pool
}

func NewWalkSpotRepository(pool *pgxpool.Pool) *WalkSpotRepository {
	return &WalkSpotRepository{pool: pool}
}

func (r *WalkSpotRepository) ListNearby(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpot, error) {
	rows, err := r.pool.Query(ctx, `
		WITH center AS (SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography AS point)
		SELECT `+walkSpotColumns+`
		FROM walk_spots, center
		WHERE ST_DWithin(location, center.point, $3)
		ORDER BY location <-> center.point
		LIMIT $4`,
		query.Center.Lng, query.Center.Lat, query.RadiusM, maxNearbySpots,
	)
	if err != nil {
		return nil, fmt.Errorf("select nearby walk spots: %w", err)
	}
	spots, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.WalkSpot, error) {
		return scanWalkSpot(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan walk spots: %w", err)
	}
	return spots, nil
}

// ListByDistance returns the spots within query.RadiusM metres, closest first, with the distance to each.
func (r *WalkSpotRepository) ListByDistance(ctx context.Context, query core_domain.NearbyQuery) ([]core_domain.WalkSpotWithDistance, error) {
	rows, err := r.pool.Query(ctx, `
		WITH center AS (SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography AS point)
		SELECT `+walkSpotColumns+`, ST_Distance(location, center.point) AS distance
		FROM walk_spots, center
		WHERE ST_DWithin(location, center.point, $3)
		ORDER BY distance, id
		LIMIT $4`,
		query.Center.Lng, query.Center.Lat, query.RadiusM, maxNearbySpots,
	)
	if err != nil {
		return nil, fmt.Errorf("select walk spots by distance: %w", err)
	}
	spots, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.WalkSpotWithDistance, error) {
		var spot core_domain.WalkSpotWithDistance
		var distance float64
		err := row.Scan(&spot.Id, &spot.Name, &spot.Location.Lat, &spot.Location.Lng, &spot.Tags, &spot.CreatedAt, &distance)
		spot.DistanceM = int(math.Round(distance))
		return spot, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan walk spots: %w", err)
	}
	return spots, nil
}

func (r *WalkSpotRepository) GetWalkSpot(ctx context.Context, id string) (core_domain.WalkSpot, error) {
	spot, err := scanWalkSpot(r.pool.QueryRow(ctx, `SELECT `+walkSpotColumns+` FROM walk_spots WHERE id = $1`, id))
	if err != nil {
		return core_domain.WalkSpot{}, fmt.Errorf("select walk spot %s: %w", id, core_postgres.MapError(err))
	}
	return spot, nil
}

func scanWalkSpot(row pgx.Row) (core_domain.WalkSpot, error) {
	var spot core_domain.WalkSpot
	err := row.Scan(&spot.Id, &spot.Name, &spot.Location.Lat, &spot.Location.Lng, &spot.Tags, &spot.CreatedAt)
	return spot, err
}
