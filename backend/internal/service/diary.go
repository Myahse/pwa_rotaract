package service

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
	"github.com/rotaract-civ/backend/internal/storage"
)

type DiaryService struct {
	diary        *repository.DiaryRepository
	clubs        *repository.ClubRepository
	users        *repository.UserRepository
	profiles     *ProfileService
	store        *storage.LocalStore
	maxImageSize int64
}

func NewDiaryService(
	diary *repository.DiaryRepository,
	clubs *repository.ClubRepository,
	users *repository.UserRepository,
	profiles *ProfileService,
	store *storage.LocalStore,
	maxImageSize int64,
) *DiaryService {
	if maxImageSize <= 0 {
		maxImageSize = defaultMaxAvatarBytes
	}
	return &DiaryService{
		diary:        diary,
		clubs:        clubs,
		users:        users,
		profiles:     profiles,
		store:        store,
		maxImageSize: maxImageSize,
	}
}

type CreateDiaryEntryInput struct {
	ParentID     *uuid.UUID                `json:"parent_id"`
	EntryType    domain.ClubDiaryEntryType `json:"entry_type"`
	UserID       *uuid.UUID                `json:"user_id"`
	FirstName    string                    `json:"first_name"`
	LastName     string                    `json:"last_name"`
	Organization *string                   `json:"organization"`
	StartedAt    *time.Time                `json:"started_at"`
	EndedAt      *time.Time                `json:"ended_at"`
	Notes        *string                   `json:"notes"`
	SortOrder    *int                      `json:"sort_order"`
}

type UpdateDiaryEntryInput struct {
	ParentID     *uuid.UUID                 `json:"parent_id"`
	EntryType    *domain.ClubDiaryEntryType `json:"entry_type"`
	UserID       *uuid.UUID                 `json:"user_id"`
	FirstName    *string                    `json:"first_name"`
	LastName     *string                    `json:"last_name"`
	Organization *string                    `json:"organization"`
	StartedAt    *time.Time                 `json:"started_at"`
	EndedAt      *time.Time                 `json:"ended_at"`
	Notes        *string                    `json:"notes"`
	SortOrder    *int                       `json:"sort_order"`
}

func (s *DiaryService) GetClubDiary(ctx context.Context, clubID uuid.UUID) (*domain.ClubDiaryResponse, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}
	entries, err := s.diary.ListByClub(ctx, clubID)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		s.enrichEntry(&entries[i])
	}
	return &domain.ClubDiaryResponse{
		Entries: entries,
		Tree:    buildDiaryTree(entries),
	}, nil
}

func (s *DiaryService) CreateEntry(ctx context.Context, actorID, clubID uuid.UUID, input CreateDiaryEntryInput) (*domain.ClubDiaryEntry, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}
	entry, err := s.prepareEntry(ctx, clubID, actorID, input)
	if err != nil {
		return nil, err
	}
	if err := s.diary.Create(ctx, entry); err != nil {
		return nil, err
	}
	s.enrichEntry(entry)
	return entry, nil
}

func (s *DiaryService) UpdateEntry(ctx context.Context, clubID, entryID uuid.UUID, input UpdateDiaryEntryInput) (*domain.ClubDiaryEntry, error) {
	entry, err := s.diary.GetByID(ctx, clubID, entryID)
	if err != nil {
		return nil, err
	}
	if input.ParentID != nil {
		if *input.ParentID != uuid.Nil {
			if err := s.validateParent(ctx, clubID, *input.ParentID, entryID); err != nil {
				return nil, err
			}
		}
		entry.ParentID = input.ParentID
	}
	if input.EntryType != nil {
		entry.EntryType = *input.EntryType
	}
	if input.UserID != nil {
		entry.UserID = input.UserID
		if *input.UserID != uuid.Nil {
			if err := s.applyUserToEntry(ctx, entry, *input.UserID); err != nil {
				return nil, err
			}
		}
	}
	if input.FirstName != nil {
		trimmed := strings.TrimSpace(*input.FirstName)
		if trimmed == "" {
			return nil, fmt.Errorf("first_name cannot be empty")
		}
		entry.FirstName = trimmed
	}
	if input.LastName != nil {
		trimmed := strings.TrimSpace(*input.LastName)
		if trimmed == "" {
			return nil, fmt.Errorf("last_name cannot be empty")
		}
		entry.LastName = trimmed
	}
	if input.Organization != nil {
		entry.Organization = input.Organization
	}
	if input.StartedAt != nil {
		entry.StartedAt = input.StartedAt
	}
	if input.EndedAt != nil {
		entry.EndedAt = input.EndedAt
	}
	if input.Notes != nil {
		entry.Notes = input.Notes
	}
	if input.SortOrder != nil {
		entry.SortOrder = *input.SortOrder
	}
	updated, err := s.diary.Update(ctx, entry)
	if err != nil {
		return nil, err
	}
	s.enrichEntry(updated)
	return updated, nil
}

func (s *DiaryService) DeleteEntry(ctx context.Context, clubID, entryID uuid.UUID) error {
	return s.diary.Delete(ctx, clubID, entryID)
}

func (s *DiaryService) UploadPhoto(
	ctx context.Context,
	clubID, entryID uuid.UUID,
	filename string,
	size int64,
	contentType string,
	file io.Reader,
) (*domain.ClubDiaryEntry, error) {
	entry, err := s.diary.GetByID(ctx, clubID, entryID)
	if err != nil {
		return nil, err
	}
	ext, err := imageExtension(filename, contentType)
	if err != nil {
		return nil, err
	}
	if size > s.maxImageSize {
		return nil, ErrImageTooLarge
	}
	limited := io.LimitReader(file, s.maxImageSize+1)
	path, err := s.store.SaveDiaryPhoto(entry.ID.String(), ext, limited)
	if err != nil {
		return nil, err
	}
	if entry.PhotoPath != nil {
		_ = s.store.RemoveByRelativePath(*entry.PhotoPath)
	}
	updated, err := s.diary.UpdatePhotoPath(ctx, clubID, entryID, path)
	if err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	s.enrichEntry(updated)
	return updated, nil
}

func (s *DiaryService) FindParrainParentID(ctx context.Context, clubID uuid.UUID) *uuid.UUID {
	entries, err := s.diary.ListByClub(ctx, clubID)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if entry.EntryType == domain.ClubDiaryEntryParrain && (entry.ParentID == nil || *entry.ParentID == uuid.Nil) {
			id := entry.ID
			return &id
		}
	}
	return nil
}

func (s *DiaryService) RecordPresidentFromUser(ctx context.Context, actorID, clubID uuid.UUID, user *domain.User, parentID *uuid.UUID, startedAt *time.Time) error {
	if user == nil {
		return nil
	}
	userID := user.ID
	entry := &domain.ClubDiaryEntry{
		ClubID:    clubID,
		ParentID:  parentID,
		EntryType: domain.ClubDiaryEntryPresident,
		UserID:    &userID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		StartedAt: startedAt,
		CreatedBy: &actorID,
	}
	if user.AvatarPath != nil {
		entry.PhotoPath = user.AvatarPath
	}
	return s.diary.Create(ctx, entry)
}

func (s *DiaryService) prepareEntry(ctx context.Context, clubID, actorID uuid.UUID, input CreateDiaryEntryInput) (*domain.ClubDiaryEntry, error) {
	if input.EntryType != domain.ClubDiaryEntryParrain &&
		input.EntryType != domain.ClubDiaryEntryPresident &&
		input.EntryType != domain.ClubDiaryEntryMember {
		return nil, fmt.Errorf("invalid entry_type")
	}
	if input.ParentID != nil && *input.ParentID != uuid.Nil {
		if err := s.validateParent(ctx, clubID, *input.ParentID, uuid.Nil); err != nil {
			return nil, err
		}
	}
	entry := &domain.ClubDiaryEntry{
		ClubID:       clubID,
		ParentID:     input.ParentID,
		EntryType:    input.EntryType,
		UserID:       input.UserID,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Organization: input.Organization,
		StartedAt:    input.StartedAt,
		EndedAt:      input.EndedAt,
		Notes:        input.Notes,
		CreatedBy:    &actorID,
	}
	if input.SortOrder != nil {
		entry.SortOrder = *input.SortOrder
	}
	if input.UserID != nil && *input.UserID != uuid.Nil {
		if err := s.applyUserToEntry(ctx, entry, *input.UserID); err != nil {
			return nil, err
		}
	}
	if entry.FirstName == "" || entry.LastName == "" {
		return nil, fmt.Errorf("first_name and last_name are required")
	}
	return entry, nil
}

func (s *DiaryService) applyUserToEntry(ctx context.Context, entry *domain.ClubDiaryEntry, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	entry.UserID = &userID
	if entry.FirstName == "" {
		entry.FirstName = user.FirstName
	}
	if entry.LastName == "" {
		entry.LastName = user.LastName
	}
	return nil
}

func (s *DiaryService) validateParent(ctx context.Context, clubID, parentID, selfID uuid.UUID) error {
	if parentID == selfID {
		return fmt.Errorf("entry cannot be its own parent")
	}
	parent, err := s.diary.GetByID(ctx, clubID, parentID)
	if err != nil {
		return fmt.Errorf("parent entry not found")
	}
	switch parent.EntryType {
	case domain.ClubDiaryEntryParrain:
		return nil
	case domain.ClubDiaryEntryPresident:
		return nil
	case domain.ClubDiaryEntryMember:
		return fmt.Errorf("members cannot be parents in the genealogy tree")
	default:
		return fmt.Errorf("invalid parent entry type")
	}
}

func (s *DiaryService) enrichEntry(entry *domain.ClubDiaryEntry) {
	entry.PhotoURL = s.profiles.UploadURL(entry.PhotoPath)
	if entry.UserID != nil {
		if user, err := s.users.GetByID(context.Background(), *entry.UserID); err == nil {
			entry.User = s.profiles.PublicUser(user)
		}
	}
}

func buildDiaryTree(entries []domain.ClubDiaryEntry) []domain.ClubDiaryTreeNode {
	childrenOf := make(map[uuid.UUID][]domain.ClubDiaryEntry)
	roots := make([]domain.ClubDiaryEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.ParentID == nil || *entry.ParentID == uuid.Nil {
			roots = append(roots, entry)
			continue
		}
		childrenOf[*entry.ParentID] = append(childrenOf[*entry.ParentID], entry)
	}

	sortEntries := func(list []domain.ClubDiaryEntry) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].SortOrder != list[j].SortOrder {
				return list[i].SortOrder < list[j].SortOrder
			}
			if list[i].StartedAt != nil && list[j].StartedAt != nil {
				return list[i].StartedAt.Before(*list[j].StartedAt)
			}
			return list[i].CreatedAt.Before(list[j].CreatedAt)
		})
	}

	var build func(entry domain.ClubDiaryEntry) domain.ClubDiaryTreeNode
	build = func(entry domain.ClubDiaryEntry) domain.ClubDiaryTreeNode {
		node := domain.ClubDiaryTreeNode{ClubDiaryEntry: entry, Children: []domain.ClubDiaryTreeNode{}}
		kids := childrenOf[entry.ID]
		sortEntries(kids)
		for _, kid := range kids {
			node.Children = append(node.Children, build(kid))
		}
		return node
	}

	sortEntries(roots)
	tree := make([]domain.ClubDiaryTreeNode, 0, len(roots))
	for _, root := range roots {
		tree = append(tree, build(root))
	}
	return tree
}
