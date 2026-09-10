package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

type PublicEventService struct {
	events *repository.PublicEventRepository
	store  *storage.LocalStore
	apiURL string
	maxSize int64
}

func NewPublicEventService(events *repository.PublicEventRepository, store *storage.LocalStore, apiPublicURL string, maxSize int64) *PublicEventService {
	if maxSize <= 0 {
		maxSize = defaultMaxAvatarBytes
	}
	return &PublicEventService{
		events:  events,
		store:   store,
		apiURL:  strings.TrimRight(apiPublicURL, "/"),
		maxSize: maxSize,
	}
}

type PublicEventInput struct {
	Published  bool      `json:"published"`
	StartsAt   time.Time `json:"starts_at"`
	City       string    `json:"city"`
	VenueFr    string    `json:"venue_fr"`
	VenueEn    string    `json:"venue_en"`
	TitleFr    string    `json:"title_fr"`
	TitleEn    string    `json:"title_en"`
	SummaryFr  string    `json:"summary_fr"`
	SummaryEn  string    `json:"summary_en"`
	BodyFr     string    `json:"body_fr"`
	BodyEn     string    `json:"body_en"`
	CtaLabelFr string    `json:"cta_label_fr"`
	CtaLabelEn string    `json:"cta_label_en"`
	CtaURL     string    `json:"cta_url"`
}

func (s *PublicEventService) List(ctx context.Context, publishedOnly bool) ([]domain.PublicEvent, error) {
	rows, err := s.events.List(ctx, publishedOnly)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PublicEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, s.public(row))
	}
	return out, nil
}

func (s *PublicEventService) Get(ctx context.Context, id uuid.UUID, publishedOnly bool) (*domain.PublicEvent, error) {
	row, err := s.events.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if publishedOnly && !row.Event.Published {
		return nil, repository.ErrNotFound
	}
	event := s.public(*row)
	return &event, nil
}

func (s *PublicEventService) Create(ctx context.Context, input PublicEventInput) (*domain.PublicEvent, error) {
	event, err := s.eventFromInput(uuid.Nil, input)
	if err != nil {
		return nil, err
	}
	if err := s.events.Create(ctx, event); err != nil {
		return nil, err
	}
	event.Images = []domain.PublicEventImage{}
	s.setStatus(event)
	return event, nil
}

func (s *PublicEventService) Update(ctx context.Context, id uuid.UUID, input PublicEventInput) (*domain.PublicEvent, error) {
	event, err := s.eventFromInput(id, input)
	if err != nil {
		return nil, err
	}
	if err := s.events.Update(ctx, event); err != nil {
		return nil, err
	}
	return s.Get(ctx, id, false)
}

func (s *PublicEventService) Delete(ctx context.Context, id uuid.UUID) error {
	row, err := s.events.Get(ctx, id)
	if err != nil {
		return err
	}
	if row.Event.FlyerPath != nil {
		_ = s.store.RemoveByRelativePath(*row.Event.FlyerPath)
	}
	for _, img := range row.Images {
		_ = s.store.RemoveByRelativePath(img.Path)
	}
	return s.events.Delete(ctx, id)
}

func (s *PublicEventService) UploadFlyer(ctx context.Context, id uuid.UUID, filename string, size int64, contentType string, file io.Reader) (*domain.PublicEvent, error) {
	if _, err := s.events.Get(ctx, id); err != nil {
		return nil, err
	}
	path, err := s.saveImage(id.String(), uuid.New().String(), filename, size, contentType, file)
	if err != nil {
		return nil, err
	}
	if old, err := s.events.FlyerPath(ctx, id); err == nil && old != nil && *old != "" && *old != path {
		_ = s.store.RemoveByRelativePath(*old)
	}
	if err := s.events.SetFlyerPath(ctx, id, path); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	return s.Get(ctx, id, false)
}

func (s *PublicEventService) AddImage(ctx context.Context, id uuid.UUID, filename string, size int64, contentType string, file io.Reader) (*domain.PublicEvent, error) {
	if _, err := s.events.Get(ctx, id); err != nil {
		return nil, err
	}
	imageID := uuid.New()
	path, err := s.saveImage(id.String(), imageID.String(), filename, size, contentType, file)
	if err != nil {
		return nil, err
	}
	if _, err := s.events.AddImage(ctx, id, path); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	return s.Get(ctx, id, false)
}

func (s *PublicEventService) DeleteImage(ctx context.Context, eventID, imageID uuid.UUID) (*domain.PublicEvent, error) {
	path, err := s.events.ImagePath(ctx, eventID, imageID)
	if err != nil {
		return nil, err
	}
	if err := s.events.DeleteImage(ctx, eventID, imageID); err != nil {
		return nil, err
	}
	_ = s.store.RemoveByRelativePath(path)
	return s.Get(ctx, eventID, false)
}

func (s *PublicEventService) saveImage(eventID, fileID, filename string, size int64, contentType string, file io.Reader) (string, error) {
	ext, err := imageExtension(filename, contentType)
	if err != nil {
		return "", err
	}
	if size > s.maxSize {
		return "", ErrImageTooLarge
	}
	return s.store.SaveEventImage(eventID, fileID, ext, io.LimitReader(file, s.maxSize+1))
}

func (s *PublicEventService) eventFromInput(id uuid.UUID, input PublicEventInput) (*domain.PublicEvent, error) {
	title := strings.TrimSpace(input.TitleFr)
	if title == "" {
		return nil, fmt.Errorf("title_fr is required")
	}
	if input.StartsAt.IsZero() {
		return nil, fmt.Errorf("starts_at is required")
	}
	event := &domain.PublicEvent{
		ID:         id,
		Published:  input.Published,
		StartsAt:   input.StartsAt,
		City:       strings.TrimSpace(input.City),
		VenueFr:    strings.TrimSpace(input.VenueFr),
		VenueEn:    strings.TrimSpace(input.VenueEn),
		TitleFr:    title,
		TitleEn:    strings.TrimSpace(input.TitleEn),
		SummaryFr:  strings.TrimSpace(input.SummaryFr),
		SummaryEn:  strings.TrimSpace(input.SummaryEn),
		BodyFr:     strings.TrimSpace(input.BodyFr),
		BodyEn:     strings.TrimSpace(input.BodyEn),
		CtaLabelFr: strings.TrimSpace(input.CtaLabelFr),
		CtaLabelEn: strings.TrimSpace(input.CtaLabelEn),
		CtaURL:     strings.TrimSpace(input.CtaURL),
	}
	s.setStatus(event)
	return event, nil
}

func (s *PublicEventService) public(row repository.PublicEventRecord) domain.PublicEvent {
	event := row.Event
	event.FlyerURL = s.uploadURL(event.FlyerPath)
	event.Images = make([]domain.PublicEventImage, 0, len(row.Images))
	for _, img := range row.Images {
		url := s.uploadURL(&img.Path)
		if url == nil {
			continue
		}
		event.Images = append(event.Images, domain.PublicEventImage{ID: img.ID, URL: *url})
	}
	s.setStatus(&event)
	return event
}

func (s *PublicEventService) setStatus(event *domain.PublicEvent) {
	if event.StartsAt.Before(time.Now()) {
		event.Status = "past"
	} else {
		event.Status = "upcoming"
	}
}

func (s *PublicEventService) uploadURL(relativePath *string) *string {
	if relativePath == nil || *relativePath == "" {
		return nil
	}
	url := s.apiURL + "/api/v1/uploads/" + strings.TrimPrefix(*relativePath, "/")
	return &url
}
