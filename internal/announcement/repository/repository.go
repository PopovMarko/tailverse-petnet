package announcement_repository

import (
	"context"
	"fmt"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxAnnouncements = 200

const announcementColumns = `a.id, a.pet_id, a.spot_id, ST_Y(a.custom_point::geometry), ST_X(a.custom_point::geometry),
	a.starts_at, a.duration_min, a.status, a.created_at`

type AnnouncementRepository struct {
	pool *pgxpool.Pool
}

func NewAnnouncementRepository(pool *pgxpool.Pool) *AnnouncementRepository {
	return &AnnouncementRepository{pool: pool}
}

func (r *AnnouncementRepository) CreateAnnouncement(ctx context.Context, a core_domain.Announcement) (core_domain.Announcement, error) {
	var lat, lng *float64
	if a.CustomPoint != nil {
		lat, lng = &a.CustomPoint.Lat, &a.CustomPoint.Lng
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO walk_announcements AS a (pet_id, spot_id, custom_point, starts_at, duration_min, status)
		VALUES (
			$1, $2,
			CASE WHEN $3::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($4::float8, $3::float8), 4326)::geography END,
			$5, $6, $7
		)
		RETURNING `+announcementColumns,
		a.PetId, a.SpotId, lat, lng, a.StartsAt, a.DurationMin, a.Status,
	)
	created, err := scanAnnouncement(row)
	if err != nil {
		return core_domain.Announcement{}, fmt.Errorf("insert announcement: %w", core_postgres.MapError(err))
	}
	return created, nil
}

func (r *AnnouncementRepository) GetAnnouncement(ctx context.Context, id string) (core_domain.Announcement, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+announcementColumns+` FROM walk_announcements a WHERE a.id = $1`, id)
	announcement, err := scanAnnouncement(row)
	if err != nil {
		return core_domain.Announcement{}, fmt.Errorf("select announcement %s: %w", id, core_postgres.MapError(err))
	}
	return announcement, nil
}

// ListAnnouncements returns active walks near the point that have not ended before "from" and start before "to".
// The place of a walk is its spot's location, or the custom point when there is no spot.
func (r *AnnouncementRepository) ListAnnouncements(ctx context.Context, filter core_domain.AnnouncementFilter, from time.Time) ([]core_domain.Announcement, error) {
	rows, err := r.pool.Query(ctx, `
		WITH center AS (SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography AS point)
		SELECT `+announcementColumns+`
		FROM walk_announcements a
		LEFT JOIN walk_spots s ON s.id = a.spot_id
		CROSS JOIN center
		WHERE a.status = 'active'
		  AND ST_DWithin(COALESCE(s.location, a.custom_point), center.point, $3)
		  AND a.starts_at + make_interval(mins => a.duration_min) >= $4
		  AND ($5::timestamptz IS NULL OR a.starts_at <= $5)
		ORDER BY a.starts_at, a.id
		LIMIT $6`,
		filter.Nearby.Center.Lng, filter.Nearby.Center.Lat, filter.Nearby.RadiusM, from, filter.To, maxAnnouncements,
	)
	if err != nil {
		return nil, fmt.Errorf("select announcements: %w", err)
	}
	announcements, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.Announcement, error) {
		return scanAnnouncement(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan announcements: %w", err)
	}
	return announcements, nil
}

func (r *AnnouncementRepository) ListParticipants(ctx context.Context, announcementId string) ([]core_domain.Participant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, CASE WHEN o.is_profile_public THEN o.nickname ELSE '' END, ap.joined_at
		FROM announcement_participants ap
		JOIN pets p ON p.id = ap.pet_id
		JOIN owners o ON o.id = p.owner_id
		WHERE ap.announcement_id = $1
		ORDER BY ap.joined_at, p.id`,
		announcementId,
	)
	if err != nil {
		return nil, fmt.Errorf("select participants: %w", err)
	}
	participants, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.Participant, error) {
		var p core_domain.Participant
		err := row.Scan(&p.PetId, &p.PetName, &p.OwnerNickname, &p.JoinedAt)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan participants: %w", err)
	}
	return participants, nil
}

func (r *AnnouncementRepository) AddParticipant(ctx context.Context, announcementId, petId string) (core_domain.Participation, error) {
	participation := core_domain.Participation{AnnouncementId: announcementId, PetId: petId}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO announcement_participants (announcement_id, pet_id)
		VALUES ($1, $2)
		RETURNING joined_at`,
		announcementId, petId,
	).Scan(&participation.JoinedAt)
	if err != nil {
		return core_domain.Participation{}, fmt.Errorf("insert participant: %w", core_postgres.MapError(err))
	}
	return participation, nil
}

func (r *AnnouncementRepository) RemoveParticipant(ctx context.Context, announcementId, petId string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM announcement_participants WHERE announcement_id = $1 AND pet_id = $2`, announcementId, petId)
	if err != nil {
		return fmt.Errorf("delete participant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pet %s has not joined announcement %s: %w", petId, announcementId, core_postgres.MapError(pgx.ErrNoRows))
	}
	return nil
}

func scanAnnouncement(row pgx.Row) (core_domain.Announcement, error) {
	var (
		a        core_domain.Announcement
		lat, lng *float64
	)
	if err := row.Scan(&a.Id, &a.PetId, &a.SpotId, &lat, &lng, &a.StartsAt, &a.DurationMin, &a.Status, &a.CreatedAt); err != nil {
		return core_domain.Announcement{}, err
	}
	if lat != nil && lng != nil {
		a.CustomPoint = &core_domain.GeoPoint{Lat: *lat, Lng: *lng}
	}
	return a, nil
}
