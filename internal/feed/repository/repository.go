package feed_repository

import (
	"context"
	"fmt"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	core_postgres "github.com/PopovMarko/tailverse-petnet/internal/core/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const postColumns = `id, pet_id, spot_id, text, photo_urls, created_at`

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

func (r *PostRepository) CreatePost(ctx context.Context, post core_domain.Post) (core_domain.Post, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO posts (pet_id, spot_id, text, photo_urls)
		VALUES ($1, $2, $3, $4)
		RETURNING `+postColumns,
		post.PetId, post.SpotId, post.Text, post.PhotoUrls,
	)
	created, err := scanPost(row)
	if err != nil {
		return core_domain.Post{}, fmt.Errorf("insert post: %w", core_postgres.MapError(err))
	}
	return created, nil
}

func (r *PostRepository) GetPost(ctx context.Context, id string) (core_domain.Post, error) {
	post, err := scanPost(r.pool.QueryRow(ctx, `SELECT `+postColumns+` FROM posts WHERE id = $1`, id))
	if err != nil {
		return core_domain.Post{}, fmt.Errorf("select post %s: %w", id, core_postgres.MapError(err))
	}
	return post, nil
}

// ListPosts returns up to filter.Limit posts, newest first, strictly after the cursor.
func (r *PostRepository) ListPosts(ctx context.Context, filter core_domain.PostFilter) ([]core_domain.Post, error) {
	var (
		afterCreatedAt any
		afterId        any
	)
	if filter.After != nil {
		afterCreatedAt, afterId = filter.After.CreatedAt, filter.After.Id
	}

	rows, err := r.pool.Query(ctx, `
		SELECT `+postColumns+`
		FROM posts
		WHERE ($1::uuid IS NULL OR spot_id = $1)
		  AND ($2::timestamptz IS NULL OR (created_at, id) < ($2::timestamptz, $3::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $4`,
		filter.SpotId, afterCreatedAt, afterId, filter.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("select posts: %w", err)
	}
	posts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (core_domain.Post, error) {
		return scanPost(row)
	})
	if err != nil {
		return nil, fmt.Errorf("scan posts: %w", err)
	}
	return posts, nil
}

func (r *PostRepository) DeletePost(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete post %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete post %s: %w", id, core_postgres.MapError(pgx.ErrNoRows))
	}
	return nil
}

func scanPost(row pgx.Row) (core_domain.Post, error) {
	var post core_domain.Post
	err := row.Scan(&post.Id, &post.PetId, &post.SpotId, &post.Text, &post.PhotoUrls, &post.CreatedAt)
	return post, err
}
