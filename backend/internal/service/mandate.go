package service

import (
	"context"
	"errors"
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

var ErrDuplicateExecutiveRole = errors.New("this executive role is already assigned for the mandate")

type MandateService struct {
	mandates     *repository.MandateRepository
	diary        *repository.DiaryRepository
	clubs        *repository.ClubRepository
	users        *repository.UserRepository
	profiles     *ProfileService
	store        *storage.LocalStore
	maxImageSize int64
}

func NewMandateService(
	mandates *repository.MandateRepository,
	diary *repository.DiaryRepository,
	clubs *repository.ClubRepository,
	users *repository.UserRepository,
	profiles *ProfileService,
	store *storage.LocalStore,
	maxImageSize int64,
) *MandateService {
	if maxImageSize <= 0 {
		maxImageSize = defaultMaxAvatarBytes
	}
	return &MandateService{
		mandates:     mandates,
		diary:        diary,
		clubs:        clubs,
		users:        users,
		profiles:     profiles,
		store:        store,
		maxImageSize: maxImageSize,
	}
}

type CreateMandateInput struct {
	Name      string     `json:"name"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	IsCurrent *bool      `json:"is_current"`
	Notes     *string    `json:"notes"`
	SortOrder *int       `json:"sort_order"`
}

type UpdateMandateInput struct {
	Name      *string    `json:"name"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	IsCurrent *bool      `json:"is_current"`
	Notes     *string    `json:"notes"`
	SortOrder *int       `json:"sort_order"`
}

type CreateMandateAssignmentInput struct {
	Role           domain.ClubMandateRole `json:"role"`
	CommissionID   *uuid.UUID             `json:"commission_id"`
	CommissionName *string                `json:"commission_name"`
	UserID         *uuid.UUID             `json:"user_id"`
	FirstName      string                 `json:"first_name"`
	LastName       string                 `json:"last_name"`
	Notes          *string                `json:"notes"`
	SortOrder      *int                   `json:"sort_order"`
}

type UpdateMandateAssignmentInput struct {
	Role           *domain.ClubMandateRole `json:"role"`
	CommissionID   *uuid.UUID              `json:"commission_id"`
	CommissionName *string                 `json:"commission_name"`
	UserID         *uuid.UUID              `json:"user_id"`
	FirstName      *string                 `json:"first_name"`
	LastName       *string                 `json:"last_name"`
	Notes          *string                 `json:"notes"`
	SortOrder      *int                    `json:"sort_order"`
}

func (s *MandateService) GetClubDiary(ctx context.Context, clubID uuid.UUID) (*domain.ClubDiaryResponse, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}

	entries, err := s.diary.ListByClub(ctx, clubID)
	if err != nil {
		return nil, err
	}
	parrains := make([]domain.ClubDiaryEntry, 0)
	legacy := make([]domain.ClubDiaryEntry, 0)
	for i := range entries {
		s.enrichDiaryEntry(&entries[i])
		if entries[i].EntryType == domain.ClubDiaryEntryParrain {
			parrains = append(parrains, entries[i])
		} else {
			legacy = append(legacy, entries[i])
		}
	}

	mandates, err := s.loadMandatesWithAssignments(ctx, clubID)
	if err != nil {
		return nil, err
	}

	return &domain.ClubDiaryResponse{
		Entries:  entries,
		Parrains: parrains,
		Mandates: mandates,
		Tree:     buildMandateDiaryTree(parrains, mandates, legacy),
	}, nil
}

func (s *MandateService) CreateMandate(ctx context.Context, actorID, clubID uuid.UUID, input CreateMandateInput) (*domain.ClubMandate, error) {
	if _, err := s.clubs.GetByID(ctx, clubID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("mandate name is required")
	}
	mandate := &domain.ClubMandate{
		ClubID:    clubID,
		Name:      name,
		StartedAt: input.StartedAt,
		EndedAt:   input.EndedAt,
		Notes:     input.Notes,
		CreatedBy: &actorID,
	}
	if input.IsCurrent != nil {
		mandate.IsCurrent = *input.IsCurrent
	}
	if input.SortOrder != nil {
		mandate.SortOrder = *input.SortOrder
	}
	if err := s.mandates.Create(ctx, mandate); err != nil {
		return nil, err
	}
	return s.mandateWithAssignments(ctx, *mandate)
}

func (s *MandateService) UpdateMandate(ctx context.Context, clubID, mandateID uuid.UUID, input UpdateMandateInput) (*domain.ClubMandate, error) {
	mandate, err := s.mandates.GetByID(ctx, clubID, mandateID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("mandate name cannot be empty")
		}
		mandate.Name = trimmed
	}
	if input.StartedAt != nil {
		mandate.StartedAt = *input.StartedAt
	}
	if input.EndedAt != nil {
		mandate.EndedAt = input.EndedAt
	}
	if input.IsCurrent != nil {
		mandate.IsCurrent = *input.IsCurrent
	}
	if input.Notes != nil {
		mandate.Notes = input.Notes
	}
	if input.SortOrder != nil {
		mandate.SortOrder = *input.SortOrder
	}
	updated, err := s.mandates.Update(ctx, mandate)
	if err != nil {
		return nil, err
	}
	return s.mandateWithAssignments(ctx, *updated)
}

func (s *MandateService) DeleteMandate(ctx context.Context, clubID, mandateID uuid.UUID) error {
	return s.mandates.Delete(ctx, clubID, mandateID)
}

func (s *MandateService) CreateAssignment(ctx context.Context, clubID, mandateID uuid.UUID, input CreateMandateAssignmentInput) (*domain.ClubMandateAssignment, error) {
	if _, err := s.mandates.GetByID(ctx, clubID, mandateID); err != nil {
		return nil, err
	}
	assignment, err := s.prepareAssignment(ctx, mandateID, input)
	if err != nil {
		return nil, err
	}
	if err := s.mandates.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}
	s.enrichAssignment(assignment)
	return assignment, nil
}

func (s *MandateService) UpdateAssignment(ctx context.Context, clubID, mandateID, assignmentID uuid.UUID, input UpdateMandateAssignmentInput) (*domain.ClubMandateAssignment, error) {
	if _, err := s.mandates.GetByID(ctx, clubID, mandateID); err != nil {
		return nil, err
	}
	assignment, err := s.mandates.GetAssignment(ctx, mandateID, assignmentID)
	if err != nil {
		return nil, err
	}
	if input.Role != nil {
		assignment.Role = *input.Role
	}
	if input.CommissionID != nil {
		assignment.CommissionID = input.CommissionID
	}
	if input.CommissionName != nil {
		assignment.CommissionName = input.CommissionName
	}
	if input.UserID != nil {
		assignment.UserID = input.UserID
		if *input.UserID != uuid.Nil {
			if err := s.applyUserToAssignment(ctx, assignment, *input.UserID); err != nil {
				return nil, err
			}
		}
	}
	if input.FirstName != nil {
		assignment.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		assignment.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Notes != nil {
		assignment.Notes = input.Notes
	}
	if input.SortOrder != nil {
		assignment.SortOrder = *input.SortOrder
	}
	if err := s.validateAssignmentRole(assignment); err != nil {
		return nil, err
	}
	if err := s.validateExecutiveRoleUnique(ctx, mandateID, assignment.Role, &assignment.ID); err != nil {
		return nil, err
	}
	updated, err := s.mandates.UpdateAssignment(ctx, assignment)
	if err != nil {
		return nil, err
	}
	s.enrichAssignment(updated)
	return updated, nil
}

func (s *MandateService) DeleteAssignment(ctx context.Context, clubID, mandateID, assignmentID uuid.UUID) error {
	if _, err := s.mandates.GetByID(ctx, clubID, mandateID); err != nil {
		return err
	}
	return s.mandates.DeleteAssignment(ctx, mandateID, assignmentID)
}

func (s *MandateService) UploadAssignmentPhoto(
	ctx context.Context,
	clubID, mandateID, assignmentID uuid.UUID,
	filename string, size int64, contentType string, file io.Reader,
) (*domain.ClubMandateAssignment, error) {
	if _, err := s.mandates.GetByID(ctx, clubID, mandateID); err != nil {
		return nil, err
	}
	assignment, err := s.mandates.GetAssignment(ctx, mandateID, assignmentID)
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
	path, err := s.store.SaveDiaryPhoto(assignment.ID.String(), ext, limited)
	if err != nil {
		return nil, err
	}
	if assignment.PhotoPath != nil {
		_ = s.store.RemoveByRelativePath(*assignment.PhotoPath)
	}
	updated, err := s.mandates.UpdateAssignmentPhoto(ctx, mandateID, assignmentID, path)
	if err != nil {
		_ = s.store.RemoveByRelativePath(path)
		return nil, err
	}
	s.enrichAssignment(updated)
	return updated, nil
}

func (s *MandateService) RecordPresidentInCurrentMandate(ctx context.Context, clubID uuid.UUID, user *domain.User) error {
	if user == nil {
		return nil
	}
	mandate, err := s.mandates.GetCurrentByClub(ctx, clubID)
	if err != nil {
		if err != repository.ErrNotFound {
			return err
		}
		now := time.Now()
		start := time.Date(now.Year(), 7, 1, 0, 0, 0, 0, time.UTC)
		if now.Month() < 7 {
			start = time.Date(now.Year()-1, 7, 1, 0, 0, 0, 0, time.UTC)
		}
		end := start.AddDate(1, 0, -1)
		name := fmt.Sprintf("%d-%d", start.Year(), end.Year())
		mandate = &domain.ClubMandate{
			ClubID:    clubID,
			Name:      name,
			StartedAt: start,
			EndedAt:   &end,
			IsCurrent: true,
		}
		if err := s.mandates.Create(ctx, mandate); err != nil {
			return err
		}
	}
	userID := user.ID
	assignment := &domain.ClubMandateAssignment{
		MandateID: mandate.ID,
		Role:      domain.ClubMandateRolePresident,
		UserID:    &userID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PhotoPath: user.AvatarPath,
	}
	return s.mandates.CreateAssignment(ctx, assignment)
}

func (s *MandateService) loadMandatesWithAssignments(ctx context.Context, clubID uuid.UUID) ([]domain.ClubMandate, error) {
	mandates, err := s.mandates.ListByClub(ctx, clubID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.ClubMandate, 0, len(mandates))
	for _, mandate := range mandates {
		filled, err := s.mandateWithAssignments(ctx, mandate)
		if err != nil {
			return nil, err
		}
		result = append(result, *filled)
	}
	return result, nil
}

func (s *MandateService) mandateWithAssignments(ctx context.Context, mandate domain.ClubMandate) (*domain.ClubMandate, error) {
	assignments, err := s.mandates.ListAssignments(ctx, mandate.ID)
	if err != nil {
		return nil, err
	}
	for i := range assignments {
		s.enrichAssignment(&assignments[i])
	}
	mandate.Bureau, mandate.Commissions, mandate.Members = groupAssignments(assignments)
	return &mandate, nil
}

func groupAssignments(assignments []domain.ClubMandateAssignment) ([]domain.ClubMandateAssignment, []domain.ClubMandateCommissionGroup, []domain.ClubMandateAssignment) {
	bureau := make([]domain.ClubMandateAssignment, 0)
	members := make([]domain.ClubMandateAssignment, 0)
	commissionMap := make(map[string]*domain.ClubMandateCommissionGroup)

	for _, a := range assignments {
		if a.Role.IsExecutiveBureau() {
			bureau = append(bureau, a)
			continue
		}
		switch a.Role {
		case domain.ClubMandateRoleMember:
			members = append(members, a)
		case domain.ClubMandateRoleCommissionPresident, domain.ClubMandateRoleCommissionSecretary, domain.ClubMandateRoleCommissionMember:
			key := commissionKey(a)
			group, ok := commissionMap[key]
			if !ok {
				name := "Commission"
				if a.CommissionName != nil && *a.CommissionName != "" {
					name = *a.CommissionName
				}
				group = &domain.ClubMandateCommissionGroup{
					CommissionID:   a.CommissionID,
					CommissionName: name,
					Members:        []domain.ClubMandateAssignment{},
				}
				commissionMap[key] = group
			}
			switch a.Role {
			case domain.ClubMandateRoleCommissionPresident:
				copyA := a
				group.President = &copyA
			case domain.ClubMandateRoleCommissionSecretary:
				copyA := a
				group.Secretary = &copyA
			default:
				group.Members = append(group.Members, a)
			}
		}
	}

	commissions := make([]domain.ClubMandateCommissionGroup, 0, len(commissionMap))
	for _, g := range commissionMap {
		commissions = append(commissions, *g)
	}
	sort.Slice(commissions, func(i, j int) bool {
		return commissions[i].CommissionName < commissions[j].CommissionName
	})
	sort.Slice(bureau, func(i, j int) bool {
		return bureauRoleOrder(bureau[i].Role) < bureauRoleOrder(bureau[j].Role)
	})
	return bureau, commissions, members
}

func commissionKey(a domain.ClubMandateAssignment) string {
	if a.CommissionID != nil {
		return a.CommissionID.String()
	}
	if a.CommissionName != nil {
		return *a.CommissionName
	}
	return "unknown"
}

func bureauRoleOrder(role domain.ClubMandateRole) int {
	switch role {
	case domain.ClubMandateRolePresident:
		return 0
	case domain.ClubMandateRolePresidentElect:
		return 1
	case domain.ClubMandateRoleImmediatePastPresident:
		return 2
	case domain.ClubMandateRoleVicePresident:
		return 3
	case domain.ClubMandateRoleSecretary:
		return 4
	case domain.ClubMandateRoleAssistantSecretary:
		return 5
	case domain.ClubMandateRoleTreasurer:
		return 6
	case domain.ClubMandateRoleAssistantTreasurer:
		return 7
	case domain.ClubMandateRoleProtocol:
		return 8
	case domain.ClubMandateRoleAssistantProtocol:
		return 9
	default:
		return 99
	}
}

func (s *MandateService) prepareAssignment(ctx context.Context, mandateID uuid.UUID, input CreateMandateAssignmentInput) (*domain.ClubMandateAssignment, error) {
	assignment := &domain.ClubMandateAssignment{
		MandateID:      mandateID,
		Role:           input.Role,
		CommissionID:   input.CommissionID,
		CommissionName: input.CommissionName,
		UserID:         input.UserID,
		FirstName:      strings.TrimSpace(input.FirstName),
		LastName:       strings.TrimSpace(input.LastName),
		Notes:          input.Notes,
	}
	if input.SortOrder != nil {
		assignment.SortOrder = *input.SortOrder
	}
	if input.UserID != nil && *input.UserID != uuid.Nil {
		if err := s.applyUserToAssignment(ctx, assignment, *input.UserID); err != nil {
			return nil, err
		}
	}
	if assignment.FirstName == "" || assignment.LastName == "" {
		return nil, fmt.Errorf("first_name and last_name are required")
	}
	if err := s.validateAssignmentRole(assignment); err != nil {
		return nil, err
	}
	if err := s.validateExecutiveRoleUnique(ctx, mandateID, assignment.Role, nil); err != nil {
		return nil, err
	}
	return assignment, nil
}

func (s *MandateService) validateAssignmentRole(a *domain.ClubMandateAssignment) error {
	switch a.Role {
	case domain.ClubMandateRoleCommissionPresident, domain.ClubMandateRoleCommissionSecretary, domain.ClubMandateRoleCommissionMember:
		hasCommission := (a.CommissionID != nil && *a.CommissionID != uuid.Nil) ||
			(a.CommissionName != nil && strings.TrimSpace(*a.CommissionName) != "")
		if !hasCommission {
			return fmt.Errorf("commission_name or commission_id is required for commission roles")
		}
	default:
		a.CommissionID = nil
		a.CommissionName = nil
	}
	return nil
}

func (s *MandateService) validateExecutiveRoleUnique(ctx context.Context, mandateID uuid.UUID, role domain.ClubMandateRole, excludeID *uuid.UUID) error {
	if !role.IsExecutiveBureau() {
		return nil
	}
	count, err := s.mandates.CountAssignmentsByRole(ctx, mandateID, role, excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateExecutiveRole
	}
	return nil
}

func (s *MandateService) applyUserToAssignment(ctx context.Context, a *domain.ClubMandateAssignment, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	a.UserID = &userID
	if a.FirstName == "" {
		a.FirstName = user.FirstName
	}
	if a.LastName == "" {
		a.LastName = user.LastName
	}
	return nil
}

func (s *MandateService) enrichAssignment(a *domain.ClubMandateAssignment) {
	a.PhotoURL = s.profiles.UploadURL(a.PhotoPath)
	if a.UserID != nil {
		if user, err := s.users.GetByID(context.Background(), *a.UserID); err == nil {
			a.User = s.profiles.PublicUser(user)
		}
	}
}

func (s *MandateService) enrichDiaryEntry(entry *domain.ClubDiaryEntry) {
	entry.PhotoURL = s.profiles.UploadURL(entry.PhotoPath)
	if entry.UserID != nil {
		if user, err := s.users.GetByID(context.Background(), *entry.UserID); err == nil {
			entry.User = s.profiles.PublicUser(user)
		}
	}
}

func buildMandateDiaryTree(parrains []domain.ClubDiaryEntry, mandates []domain.ClubMandate, legacy []domain.ClubDiaryEntry) []domain.ClubDiaryTreeNode {
	tree := make([]domain.ClubDiaryTreeNode, 0)

	for _, p := range parrains {
		node := domain.ClubDiaryTreeNode{ClubDiaryEntry: p, Children: []domain.ClubDiaryTreeNode{}}
		for _, m := range mandates {
			node.Children = append(node.Children, mandateToTreeNode(m))
		}
		tree = append(tree, node)
	}

	if len(parrains) == 0 {
		for _, m := range mandates {
			tree = append(tree, mandateToTreeNode(m))
		}
	}

	if len(tree) == 0 && len(legacy) > 0 {
		return buildDiaryTree(legacy)
	}
	return tree
}

func mandateToTreeNode(m domain.ClubMandate) domain.ClubDiaryTreeNode {
	label := m.Name
	if m.IsCurrent {
		label += " (actuel)"
	}
	node := domain.ClubDiaryTreeNode{
		ClubDiaryEntry: domain.ClubDiaryEntry{
			ID:        m.ID,
			ClubID:    m.ClubID,
			EntryType: domain.ClubDiaryEntryPresident,
			FirstName: label,
			LastName:  "Mandat",
			StartedAt: &m.StartedAt,
			EndedAt:   m.EndedAt,
			Notes:     m.Notes,
		},
		Children: []domain.ClubDiaryTreeNode{},
	}

	for _, b := range m.Bureau {
		node.Children = append(node.Children, assignmentToTreeNode(b))
	}
	for _, c := range m.Commissions {
		commNode := domain.ClubDiaryTreeNode{
			ClubDiaryEntry: domain.ClubDiaryEntry{
				ID:           uuid.New(),
				ClubID:       m.ClubID,
				EntryType:    domain.ClubDiaryEntryMember,
				FirstName:    c.CommissionName,
				LastName:     "Commission",
			},
			Children: []domain.ClubDiaryTreeNode{},
		}
		if c.President != nil {
			commNode.Children = append(commNode.Children, assignmentToTreeNode(*c.President))
		}
		if c.Secretary != nil {
			commNode.Children = append(commNode.Children, assignmentToTreeNode(*c.Secretary))
		}
		for _, member := range c.Members {
			commNode.Children = append(commNode.Children, assignmentToTreeNode(member))
		}
		node.Children = append(node.Children, commNode)
	}
	for _, member := range m.Members {
		node.Children = append(node.Children, assignmentToTreeNode(member))
	}
	return node
}

func assignmentToTreeNode(a domain.ClubMandateAssignment) domain.ClubDiaryTreeNode {
	entryType := domain.ClubDiaryEntryMember
	if a.Role.IsPresidentTier() {
		entryType = domain.ClubDiaryEntryPresident
	}
	return domain.ClubDiaryTreeNode{
		ClubDiaryEntry: domain.ClubDiaryEntry{
			ID:        a.ID,
			EntryType: entryType,
			FirstName: a.FirstName,
			LastName:  a.LastName,
			PhotoURL:  a.PhotoURL,
			Notes:     a.Notes,
		},
		Children: []domain.ClubDiaryTreeNode{},
	}
}
