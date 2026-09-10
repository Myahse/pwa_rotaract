package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

type SiteContentService struct {
	repo         *repository.SiteContentRepository
	store        *storage.LocalStore
	apiPublicURL string
	maxSize      int64
}

func NewSiteContentService(repo *repository.SiteContentRepository, store *storage.LocalStore, apiPublicURL string, maxSize int64) *SiteContentService {
	return &SiteContentService{
		repo:         repo,
		store:        store,
		apiPublicURL: strings.TrimRight(apiPublicURL, "/"),
		maxSize:      maxSize,
	}
}

type SiteGalleryInput struct {
	Published bool   `json:"published"`
	CaptionFr string `json:"caption_fr"`
	CaptionEn string `json:"caption_en"`
	SortOrder *int   `json:"sort_order"`
}

type FeaturedPostulantInput struct {
	PeriodYear   int      `json:"period_year"`
	PeriodMonth  int      `json:"period_month"`
	Published    bool     `json:"published"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	HomeClub     string   `json:"home_club"`
	QuoteFr      string   `json:"quote_fr"`
	QuoteEn      string   `json:"quote_en"`
	VisitCount   int      `json:"visit_count"`
	ClubsVisited []string `json:"clubs_visited"`
}

func previousCalendarMonth(now time.Time) (year int, month int) {
	prev := now.AddDate(0, -1, 0)
	return prev.Year(), int(prev.Month())
}

func validatePostulantPeriod(year, month int) error {
	if year < 2000 || year > 2100 {
		return errors.New("period_year is invalid")
	}
	if month < 1 || month > 12 {
		return errors.New("period_month must be between 1 and 12")
	}
	return nil
}

func applyFeaturedPostulantInput(item *domain.FeaturedPostulant, input FeaturedPostulantInput) error {
	if err := validatePostulantPeriod(input.PeriodYear, input.PeriodMonth); err != nil {
		return err
	}
	if input.VisitCount < 0 {
		return errors.New("visit_count must be zero or greater")
	}
	item.PeriodYear = input.PeriodYear
	item.PeriodMonth = input.PeriodMonth
	item.Published = input.Published
	item.FirstName = strings.TrimSpace(input.FirstName)
	item.LastName = strings.TrimSpace(input.LastName)
	item.HomeClub = strings.TrimSpace(input.HomeClub)
	item.QuoteFr = strings.TrimSpace(input.QuoteFr)
	item.QuoteEn = strings.TrimSpace(input.QuoteEn)
	item.VisitCount = input.VisitCount
	item.ClubsVisited = normalizeClubList(input.ClubsVisited)
	return nil
}

func (s *SiteContentService) ListGallery(ctx context.Context, publishedOnly bool) ([]domain.SiteGalleryImage, error) {
	items, err := s.repo.ListGallery(ctx, publishedOnly)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SiteGalleryImage, 0, len(items))
	for _, item := range items {
		out = append(out, s.publicGallery(item))
	}
	return out, nil
}

func (s *SiteContentService) CreateGallery(ctx context.Context, input SiteGalleryInput) (*domain.SiteGalleryImage, error) {
	item := &domain.SiteGalleryImage{
		Published: input.Published,
		CaptionFr: strings.TrimSpace(input.CaptionFr),
		CaptionEn: strings.TrimSpace(input.CaptionEn),
	}
	if err := s.repo.CreateGallery(ctx, item); err != nil {
		return nil, err
	}
	out := s.publicGallery(*item)
	return &out, nil
}

func (s *SiteContentService) UpdateGallery(ctx context.Context, id uuid.UUID, input SiteGalleryInput) (*domain.SiteGalleryImage, error) {
	current, err := s.repo.GetGallery(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Published = input.Published
	current.CaptionFr = strings.TrimSpace(input.CaptionFr)
	current.CaptionEn = strings.TrimSpace(input.CaptionEn)
	if input.SortOrder != nil {
		current.SortOrder = *input.SortOrder
	}
	if err := s.repo.UpdateGallery(ctx, current); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetGallery(ctx, id)
	if err != nil {
		return nil, err
	}
	out := s.publicGallery(*updated)
	return &out, nil
}

func (s *SiteContentService) UploadGalleryImage(ctx context.Context, id uuid.UUID, filename string, size int64, contentType string, file io.Reader) (*domain.SiteGalleryImage, error) {
	if _, err := s.repo.GetGallery(ctx, id); err != nil {
		return nil, err
	}
	path, err := s.saveImage(fmt.Sprintf("gallery_%s", id.String()), filename, size, contentType, file)
	if err != nil {
		return nil, err
	}
	oldPath, _ := s.repo.GalleryPath(ctx, id)
	if oldPath != "" && oldPath != path {
		_ = s.store.RemoveByRelativePath(oldPath)
	}
	if err := s.repo.SetGalleryPath(ctx, id, path); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	updated, err := s.repo.GetGallery(ctx, id)
	if err != nil {
		return nil, err
	}
	out := s.publicGallery(*updated)
	return &out, nil
}

func (s *SiteContentService) DeleteGallery(ctx context.Context, id uuid.UUID) error {
	path, err := s.repo.GalleryPath(ctx, id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if err := s.repo.DeleteGallery(ctx, id); err != nil {
		return err
	}
	if path != "" {
		_ = s.store.RemoveByRelativePath(path)
	}
	return nil
}

func (s *SiteContentService) GetFeaturedPostulantForDisplay(ctx context.Context, publishedOnly bool) (*domain.FeaturedPostulant, error) {
	year, month := previousCalendarMonth(time.Now())
	item, err := s.repo.GetFeaturedPostulantByPeriod(ctx, year, month)
	if err != nil {
		return nil, err
	}
	if publishedOnly && !item.Published {
		return nil, repository.ErrNotFound
	}
	out := s.publicPostulant(*item)
	return &out, nil
}

func (s *SiteContentService) ListFeaturedPostulants(ctx context.Context) ([]domain.FeaturedPostulant, error) {
	items, err := s.repo.ListFeaturedPostulants(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.FeaturedPostulant, 0, len(items))
	for _, item := range items {
		out = append(out, s.publicPostulant(item))
	}
	return out, nil
}

func (s *SiteContentService) CreateFeaturedPostulant(ctx context.Context, input FeaturedPostulantInput) (*domain.FeaturedPostulant, error) {
	item := &domain.FeaturedPostulant{}
	if err := applyFeaturedPostulantInput(item, input); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetFeaturedPostulantByPeriod(ctx, item.PeriodYear, item.PeriodMonth); err == nil {
		return nil, errors.New("a postulant already exists for this month")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if err := s.repo.CreateFeaturedPostulant(ctx, item); err != nil {
		return nil, err
	}
	out := s.publicPostulant(*item)
	return &out, nil
}

func (s *SiteContentService) UpdateFeaturedPostulant(ctx context.Context, id uuid.UUID, input FeaturedPostulantInput) (*domain.FeaturedPostulant, error) {
	current, err := s.repo.GetFeaturedPostulantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := applyFeaturedPostulantInput(current, input); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetFeaturedPostulantByPeriod(ctx, current.PeriodYear, current.PeriodMonth)
	if err == nil && existing.ID != id {
		return nil, errors.New("a postulant already exists for this month")
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if err := s.repo.UpdateFeaturedPostulant(ctx, current); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetFeaturedPostulantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := s.publicPostulant(*updated)
	return &out, nil
}

func (s *SiteContentService) DeleteFeaturedPostulant(ctx context.Context, id uuid.UUID) error {
	path, err := s.repo.PostulantFlyerPath(ctx, id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if err := s.repo.DeleteFeaturedPostulant(ctx, id); err != nil {
		return err
	}
	if path != nil && *path != "" {
		_ = s.store.RemoveByRelativePath(*path)
	}
	return nil
}

func (s *SiteContentService) UploadPostulantFlyer(ctx context.Context, id uuid.UUID, filename string, size int64, contentType string, file io.Reader) (*domain.FeaturedPostulant, error) {
	if _, err := s.repo.GetFeaturedPostulantByID(ctx, id); err != nil {
		return nil, err
	}
	path, err := s.saveImage("postulant_"+uuid.New().String(), filename, size, contentType, file)
	if err != nil {
		return nil, err
	}
	oldPath, _ := s.repo.PostulantFlyerPath(ctx, id)
	if oldPath != nil && *oldPath != "" && *oldPath != path {
		_ = s.store.RemoveByRelativePath(*oldPath)
	}
	if err := s.repo.SetPostulantFlyer(ctx, id, path); err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	updated, err := s.repo.GetFeaturedPostulantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := s.publicPostulant(*updated)
	return &out, nil
}

func (s *SiteContentService) saveImage(name, filename string, size int64, contentType string, file io.Reader) (string, error) {
	if size > s.maxSize {
		return "", ErrImageTooLarge
	}
	ext, err := imageExtension(filename, contentType)
	if err != nil {
		return "", err
	}
	limited := io.LimitReader(file, s.maxSize+1)
	return s.store.SaveSiteImage(name, ext, limited)
}

func (s *SiteContentService) publicGallery(item domain.SiteGalleryImage) domain.SiteGalleryImage {
	if item.Path != "" {
		url := s.uploadURL(item.Path)
		item.URL = url
	}
	item.Path = ""
	return item
}

func (s *SiteContentService) publicPostulant(item domain.FeaturedPostulant) domain.FeaturedPostulant {
	if item.FlyerPath != nil && *item.FlyerPath != "" {
		url := s.uploadURL(*item.FlyerPath)
		item.FlyerURL = &url
	}
	item.FlyerPath = nil
	return item
}

func (s *SiteContentService) uploadURL(relativePath string) string {
	clean := strings.ReplaceAll(strings.TrimPrefix(relativePath, "/"), "\\", "/")
	return "/api/v1/uploads/" + clean
}

func normalizeClubList(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}
