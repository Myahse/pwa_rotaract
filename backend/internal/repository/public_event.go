package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotaract-civ/backend/internal/domain"
)

type PublicEventRepository struct {
	pool *pgxpool.Pool
}

func NewPublicEventRepository(pool *pgxpool.Pool) *PublicEventRepository {
	return &PublicEventRepository{pool: pool}
}

type PublicEventRecord struct {
	Event  domain.PublicEvent
	Images []eventImageRow
}

type eventImageRow struct {
	ID   uuid.UUID
	Path string
}

func (r *PublicEventRepository) List(ctx context.Context, publishedOnly bool) ([]PublicEventRecord, error) {
	query := `
		SELECT id, published, starts_at, city, venue_fr, venue_en, title_fr, title_en,
		       summary_fr, summary_en, body_fr, body_en, cta_label_fr, cta_label_en, cta_url,
		       flyer_path, created_at, updated_at
		FROM public_events
	`
	if publishedOnly {
		query += " WHERE published = TRUE"
	}
	query += " ORDER BY starts_at DESC"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list public events: %w", err)
	}
	defer rows.Close()

	items := make([]PublicEventRecord, 0)
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		rec, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, PublicEventRecord{Event: rec})
		ids = append(ids, rec.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	images, err := r.listImages(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Images = images[items[i].Event.ID]
	}
	return items, nil
}

func (r *PublicEventRepository) Get(ctx context.Context, id uuid.UUID) (*PublicEventRecord, error) {
	rec, err := scanEvent(r.pool.QueryRow(ctx, `
		SELECT id, published, starts_at, city, venue_fr, venue_en, title_fr, title_en,
		       summary_fr, summary_en, body_fr, body_en, cta_label_fr, cta_label_en, cta_url,
		       flyer_path, created_at, updated_at
		FROM public_events WHERE id = $1
	`, id))
	if err != nil {
		return nil, err
	}
	images, err := r.listImages(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	return &PublicEventRecord{Event: rec, Images: images[id]}, nil
}

func (r *PublicEventRepository) Create(ctx context.Context, event *domain.PublicEvent) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO public_events (
			published, starts_at, city, venue_fr, venue_en, title_fr, title_en,
			summary_fr, summary_en, body_fr, body_en, cta_label_fr, cta_label_en, cta_url
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, updated_at
	`, event.Published, event.StartsAt, event.City, event.VenueFr, event.VenueEn, event.TitleFr, event.TitleEn,
		event.SummaryFr, event.SummaryEn, event.BodyFr, event.BodyEn, event.CtaLabelFr, event.CtaLabelEn, event.CtaURL,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *PublicEventRepository) Update(ctx context.Context, event *domain.PublicEvent) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE public_events SET
			published = $2, starts_at = $3, city = $4, venue_fr = $5, venue_en = $6,
			title_fr = $7, title_en = $8, summary_fr = $9, summary_en = $10,
			body_fr = $11, body_en = $12, cta_label_fr = $13, cta_label_en = $14, cta_url = $15,
			updated_at = NOW()
		WHERE id = $1
	`, event.ID, event.Published, event.StartsAt, event.City, event.VenueFr, event.VenueEn,
		event.TitleFr, event.TitleEn, event.SummaryFr, event.SummaryEn,
		event.BodyFr, event.BodyEn, event.CtaLabelFr, event.CtaLabelEn, event.CtaURL)
	if err != nil {
		return fmt.Errorf("update public event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PublicEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM public_events WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete public event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PublicEventRepository) SetFlyerPath(ctx context.Context, id uuid.UUID, path string) error {
	_, err := r.pool.Exec(ctx, `UPDATE public_events SET flyer_path = $2, updated_at = NOW() WHERE id = $1`, id, path)
	return err
}

func (r *PublicEventRepository) FlyerPath(ctx context.Context, id uuid.UUID) (*string, error) {
	var path *string
	err := r.pool.QueryRow(ctx, `SELECT flyer_path FROM public_events WHERE id = $1`, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return path, err
}

func (r *PublicEventRepository) AddImage(ctx context.Context, eventID uuid.UUID, path string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO public_event_images (event_id, path, sort_order)
		VALUES ($1, $2, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM public_event_images WHERE event_id = $1))
		RETURNING id
	`, eventID, path).Scan(&id)
	return id, err
}

func (r *PublicEventRepository) ImagePath(ctx context.Context, eventID, imageID uuid.UUID) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `
		SELECT path FROM public_event_images WHERE id = $1 AND event_id = $2
	`, imageID, eventID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func (r *PublicEventRepository) DeleteImage(ctx context.Context, eventID, imageID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM public_event_images WHERE id = $1 AND event_id = $2`, imageID, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PublicEventRepository) listImages(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]eventImageRow, error) {
	out := make(map[uuid.UUID][]eventImageRow)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, event_id, path FROM public_event_images
		WHERE event_id = ANY($1)
		ORDER BY sort_order, created_at
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("list event images: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, eventID uuid.UUID
		var path string
		if err := rows.Scan(&id, &eventID, &path); err != nil {
			return nil, err
		}
		out[eventID] = append(out[eventID], eventImageRow{ID: id, Path: path})
	}
	return out, rows.Err()
}

func scanEvent(row interface{ Scan(dest ...any) error }) (domain.PublicEvent, error) {
	var event domain.PublicEvent
	err := row.Scan(
		&event.ID, &event.Published, &event.StartsAt, &event.City, &event.VenueFr, &event.VenueEn,
		&event.TitleFr, &event.TitleEn, &event.SummaryFr, &event.SummaryEn, &event.BodyFr, &event.BodyEn,
		&event.CtaLabelFr, &event.CtaLabelEn, &event.CtaURL, &event.FlyerPath, &event.CreatedAt, &event.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return event, ErrNotFound
	}
	if err != nil {
		return event, fmt.Errorf("scan public event: %w", err)
	}
	if event.StartsAt.Before(time.Now()) {
		event.Status = "past"
	} else {
		event.Status = "upcoming"
	}
	if event.Images == nil {
		event.Images = []domain.PublicEventImage{}
	}
	return event, nil
}
