package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type SiteContentRepository struct {
	pool *pgxpool.Pool
}

func NewSiteContentRepository(pool *pgxpool.Pool) *SiteContentRepository {
	return &SiteContentRepository{pool: pool}
}

func (r *SiteContentRepository) ListGallery(ctx context.Context, publishedOnly bool) ([]domain.SiteGalleryImage, error) {
	query := `
		SELECT id, published, caption_fr, caption_en, path, sort_order, created_at, updated_at
		FROM site_gallery_images
	`
	if publishedOnly {
		query += ` WHERE published = TRUE AND path <> ''`
	}
	query += ` ORDER BY sort_order ASC, created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SiteGalleryImage, 0)
	for rows.Next() {
		item, err := scanGalleryImage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SiteContentRepository) GetGallery(ctx context.Context, id uuid.UUID) (*domain.SiteGalleryImage, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, published, caption_fr, caption_en, path, sort_order, created_at, updated_at
		FROM site_gallery_images WHERE id = $1
	`, id)
	item, err := scanGalleryImage(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SiteContentRepository) CreateGallery(ctx context.Context, item *domain.SiteGalleryImage) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO site_gallery_images (published, caption_fr, caption_en, path, sort_order)
		VALUES ($1, $2, $3, $4, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM site_gallery_images))
		RETURNING id, sort_order, created_at, updated_at
	`, item.Published, item.CaptionFr, item.CaptionEn, item.Path).Scan(
		&item.ID, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
	)
}

func (r *SiteContentRepository) UpdateGallery(ctx context.Context, item *domain.SiteGalleryImage) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE site_gallery_images SET
			published = $2,
			caption_fr = $3,
			caption_en = $4,
			sort_order = $5,
			updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.Published, item.CaptionFr, item.CaptionEn, item.SortOrder)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) SetGalleryPath(ctx context.Context, id uuid.UUID, path string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE site_gallery_images SET path = $2, updated_at = NOW() WHERE id = $1
	`, id, path)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) GalleryPath(ctx context.Context, id uuid.UUID) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `SELECT path FROM site_gallery_images WHERE id = $1`, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func (r *SiteContentRepository) DeleteGallery(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM site_gallery_images WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) ListFeaturedPostulants(ctx context.Context) ([]domain.FeaturedPostulant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, period_year, period_month, published, first_name, last_name, home_club, quote_fr, quote_en,
			visit_count, clubs_visited, flyer_path, updated_at
		FROM featured_postulant
		ORDER BY period_year DESC, period_month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.FeaturedPostulant, 0)
	for rows.Next() {
		item, err := scanFeaturedPostulant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SiteContentRepository) GetFeaturedPostulantByID(ctx context.Context, id uuid.UUID) (*domain.FeaturedPostulant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, period_year, period_month, published, first_name, last_name, home_club, quote_fr, quote_en,
			visit_count, clubs_visited, flyer_path, updated_at
		FROM featured_postulant WHERE id = $1
	`, id)
	item, err := scanFeaturedPostulant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SiteContentRepository) GetFeaturedPostulantByPeriod(ctx context.Context, year, month int) (*domain.FeaturedPostulant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, period_year, period_month, published, first_name, last_name, home_club, quote_fr, quote_en,
			visit_count, clubs_visited, flyer_path, updated_at
		FROM featured_postulant
		WHERE period_year = $1 AND period_month = $2
	`, year, month)
	item, err := scanFeaturedPostulant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SiteContentRepository) CreateFeaturedPostulant(ctx context.Context, item *domain.FeaturedPostulant) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO featured_postulant (
			id, period_year, period_month, published, first_name, last_name, home_club,
			quote_fr, quote_en, visit_count, clubs_visited, flyer_path
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING updated_at
	`, item.ID, item.PeriodYear, item.PeriodMonth, item.Published, item.FirstName, item.LastName, item.HomeClub,
		item.QuoteFr, item.QuoteEn, item.VisitCount, item.ClubsVisited, item.FlyerPath,
	).Scan(&item.UpdatedAt)
}

func (r *SiteContentRepository) UpdateFeaturedPostulant(ctx context.Context, item *domain.FeaturedPostulant) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE featured_postulant SET
			period_year = $2,
			period_month = $3,
			published = $4,
			first_name = $5,
			last_name = $6,
			home_club = $7,
			quote_fr = $8,
			quote_en = $9,
			visit_count = $10,
			clubs_visited = $11,
			updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.PeriodYear, item.PeriodMonth, item.Published, item.FirstName, item.LastName, item.HomeClub,
		item.QuoteFr, item.QuoteEn, item.VisitCount, item.ClubsVisited)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) DeleteFeaturedPostulant(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM featured_postulant WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) SetPostulantFlyer(ctx context.Context, id uuid.UUID, path string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE featured_postulant SET flyer_path = $2, updated_at = NOW() WHERE id = $1
	`, id, path)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SiteContentRepository) PostulantFlyerPath(ctx context.Context, id uuid.UUID) (*string, error) {
	var path *string
	err := r.pool.QueryRow(ctx, `SELECT flyer_path FROM featured_postulant WHERE id = $1`, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return path, err
}

func scanGalleryImage(row pgx.Row) (domain.SiteGalleryImage, error) {
	var item domain.SiteGalleryImage
	err := row.Scan(
		&item.ID, &item.Published, &item.CaptionFr, &item.CaptionEn, &item.Path,
		&item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func scanFeaturedPostulant(row pgx.Row) (domain.FeaturedPostulant, error) {
	var item domain.FeaturedPostulant
	var updatedAt time.Time
	err := row.Scan(
		&item.ID, &item.PeriodYear, &item.PeriodMonth, &item.Published, &item.FirstName, &item.LastName, &item.HomeClub,
		&item.QuoteFr, &item.QuoteEn, &item.VisitCount, &item.ClubsVisited,
		&item.FlyerPath, &updatedAt,
	)
	item.UpdatedAt = updatedAt
	if item.ClubsVisited == nil {
		item.ClubsVisited = []string{}
	}
	return item, err
}
