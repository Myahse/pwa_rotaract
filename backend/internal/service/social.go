package service

import (
	"context"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

var (
	ErrSocialPostNotFound  = errors.New("post not found")
	ErrSocialForbidden     = errors.New("not allowed")
	ErrSocialLoginRequired = errors.New("login required")
	ErrInvalidSocialMedia  = errors.New("media must be jpeg, png, webp, mp4, or webm")
	ErrSocialMediaTooLarge = errors.New("media file is too large")
)

const defaultMaxSocialMediaBytes int64 = 50 << 20 // 50MB

type SocialService struct {
	social   *repository.SocialRepository
	clubs    *repository.ClubRepository
	profiles *ProfileService
	files    *storage.LocalStore
	maxMedia int64
}

func NewSocialService(
	social *repository.SocialRepository,
	clubs *repository.ClubRepository,
	profiles *ProfileService,
	files *storage.LocalStore,
	maxMedia int64,
) *SocialService {
	if maxMedia <= 0 {
		maxMedia = defaultMaxSocialMediaBytes
	}
	return &SocialService{social: social, clubs: clubs, profiles: profiles, files: files, maxMedia: maxMedia}
}

type CreateSocialPostInput struct {
	Body   string
	ClubID *uuid.UUID
	Media  []SocialMediaUpload
}

type SocialMediaUpload struct {
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type ReactInput struct {
	Kind string `json:"kind"`
}

func (s *SocialService) viewerID(user *domain.User) uuid.UUID {
	if user == nil {
		return uuid.Nil
	}
	return user.ID
}

func (s *SocialService) requireUser(user *domain.User) error {
	if user == nil {
		return ErrSocialLoginRequired
	}
	return nil
}

func (s *SocialService) CreatePost(ctx context.Context, user *domain.User, input CreateSocialPostInput) (*domain.SocialPost, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	body := strings.TrimSpace(input.Body)
	if utf8.RuneCountInString(body) > 5000 {
		return nil, errors.New("body must be at most 5000 characters")
	}
	if body == "" && len(input.Media) == 0 {
		return nil, errors.New("ajoutez un texte, une photo ou une vidéo")
	}
	if len(input.Media) > 6 {
		return nil, errors.New("maximum 6 médias par publication")
	}
	if input.ClubID != nil {
		_, err := s.clubs.GetMembership(ctx, *input.ClubID, user.ID)
		if errors.Is(err, repository.ErrNotFound) {
			if !user.IsAdmin {
				return nil, errors.New("you must be a member of the selected club")
			}
		} else if err != nil {
			return nil, err
		}
	}

	post := &domain.SocialPost{
		AuthorID: user.ID,
		ClubID:   input.ClubID,
		Body:     body,
	}
	if err := s.social.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	for i, file := range input.Media {
		kind, ext, err := socialMediaKind(file.Filename, file.ContentType)
		if err != nil {
			return nil, err
		}
		if file.Size > s.maxMedia {
			return nil, ErrSocialMediaTooLarge
		}
		mediaID := uuid.New()
		path, err := s.files.SaveSocialMedia(post.ID.String(), mediaID.String(), ext, file.Reader)
		if err != nil {
			return nil, err
		}
		media := &domain.SocialPostMedia{
			PostID:    post.ID,
			Kind:      kind,
			Path:      path,
			SortOrder: i,
		}
		if err := s.social.AddMedia(ctx, media); err != nil {
			return nil, err
		}
	}

	return s.GetPost(ctx, user, post.ID)
}

func (s *SocialService) Repost(ctx context.Context, user *domain.User, postID uuid.UUID, quoteBody string) (*domain.SocialPost, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	quoteBody = strings.TrimSpace(quoteBody)
	if utf8.RuneCountInString(quoteBody) > 5000 {
		return nil, errors.New("quote must be at most 5000 characters")
	}
	post, err := s.social.CreateRepost(ctx, user.ID, postID, quoteBody)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetPost(ctx, user, post.ID)
}

func (s *SocialService) GetPost(ctx context.Context, user *domain.User, postID uuid.UUID) (*domain.SocialPost, error) {
	row, err := s.social.GetPost(ctx, postID, s.viewerID(user))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialPostNotFound
	}
	if err != nil {
		return nil, err
	}
	posts, err := s.attachMedia(ctx, []repository.SocialPostRow{*row})
	if err != nil {
		return nil, err
	}
	return &posts[0], nil
}

func (s *SocialService) ListFeed(ctx context.Context, user *domain.User, limit int, before *time.Time, followingOnly bool) ([]domain.SocialPost, error) {
	if followingOnly {
		if err := s.requireUser(user); err != nil {
			return nil, err
		}
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := s.social.ListFeed(ctx, s.viewerID(user), limit, before, followingOnly)
	if err != nil {
		return nil, err
	}
	return s.attachMedia(ctx, rows)
}

func (s *SocialService) DeletePost(ctx context.Context, user *domain.User, postID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	err := s.social.DeletePost(ctx, postID, user.ID, user.IsAdmin)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialForbidden
	}
	return err
}

func (s *SocialService) ListComments(ctx context.Context, _ *domain.User, postID uuid.UUID) ([]domain.SocialComment, error) {
	ok, err := s.social.PostExistsVisible(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialPostNotFound
	}
	comments, err := s.social.ListComments(ctx, postID)
	if err != nil {
		return nil, err
	}
	for i := range comments {
		if comments[i].Author != nil {
			comments[i].Author.AvatarURL = s.profiles.UploadURL(comments[i].Author.AvatarURL)
		}
		for j := range comments[i].Replies {
			if comments[i].Replies[j].Author != nil {
				comments[i].Replies[j].Author.AvatarURL = s.profiles.UploadURL(comments[i].Replies[j].Author.AvatarURL)
			}
		}
	}
	return comments, nil
}

func (s *SocialService) React(ctx context.Context, user *domain.User, postID uuid.UUID, input ReactInput) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	ok, err := s.social.PostExistsVisible(ctx, postID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSocialPostNotFound
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		kind = "like"
	}
	if kind != "like" {
		return errors.New("unsupported reaction")
	}
	return s.social.AddReaction(ctx, postID, user.ID, kind)
}

func (s *SocialService) Unreact(ctx context.Context, user *domain.User, postID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	err := s.social.RemoveReaction(ctx, postID, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	return err
}

func (s *SocialService) Follow(ctx context.Context, user *domain.User, targetID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	if user.ID == targetID {
		return errors.New("cannot follow yourself")
	}
	return s.social.Follow(ctx, user.ID, targetID)
}

func (s *SocialService) Unfollow(ctx context.Context, user *domain.User, targetID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	return s.social.Unfollow(ctx, user.ID, targetID)
}

func (s *SocialService) ListFollowing(ctx context.Context, user *domain.User) ([]domain.SocialFollowProfile, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	profiles, err := s.social.ListFollowing(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return s.hydrateFollowProfiles(profiles), nil
}

func (s *SocialService) Suggestions(ctx context.Context, user *domain.User) ([]domain.SocialFollowProfile, error) {
	var profiles []domain.SocialFollowProfile
	var err error
	if user == nil {
		profiles, err = s.social.PublicSuggestions(ctx, 12)
	} else {
		profiles, err = s.social.ListSuggestions(ctx, user.ID, 12)
	}
	if err != nil {
		return nil, err
	}
	return s.hydrateFollowProfiles(profiles), nil
}

func (s *SocialService) hydrateFollowProfiles(profiles []domain.SocialFollowProfile) []domain.SocialFollowProfile {
	for i := range profiles {
		profiles[i].AvatarURL = s.profiles.UploadURL(profiles[i].AvatarURL)
	}
	return profiles
}

func (s *SocialService) attachMedia(ctx context.Context, rows []repository.SocialPostRow) ([]domain.SocialPost, error) {
	if len(rows) == 0 {
		return []domain.SocialPost{}, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].Post.ID)
	}
	mediaMap, err := s.social.ListMedia(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SocialPost, 0, len(rows))
	for i := range rows {
		post := rows[i].Post
		if post.Author != nil {
			post.Author.AvatarURL = s.profiles.UploadURL(rows[i].AuthorAvatar)
		}
		media := mediaMap[post.ID]
		for j := range media {
			if url := s.profiles.UploadURL(&media[j].Path); url != nil {
				media[j].URL = *url
			}
		}
		post.Media = media
		if post.RepostedPostID != nil {
			original, err := s.GetPost(ctx, nil, *post.RepostedPostID)
			if err != nil {
				return nil, err
			}
			post.Original = original
		}
		out = append(out, post)
	}
	return out, nil
}

func socialMediaKind(filename, contentType string) (domain.SocialMediaKind, string, error) {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}
	switch contentType {
	case "image/jpeg", "image/jpg":
		return domain.SocialMediaImage, ".jpg", nil
	case "image/png":
		return domain.SocialMediaImage, ".png", nil
	case "image/webp":
		return domain.SocialMediaImage, ".webp", nil
	case "video/mp4":
		return domain.SocialMediaVideo, ".mp4", nil
	case "video/webm":
		return domain.SocialMediaVideo, ".webm", nil
	default:
		return "", "", ErrInvalidSocialMedia
	}
}
