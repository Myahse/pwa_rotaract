package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/repository"
)

var (
	ErrSocialGroupNotFound = errors.New("groupe introuvable")
	ErrSocialGroupForbidden = errors.New("action non autorisée")
	ErrSocialGroupNotMember = errors.New("vous n'êtes pas membre de ce groupe")
)

type CreateSocialGroupInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
}

type JoinSocialGroupInput struct {
	Message string `json:"message"`
}

type SendGroupMessageInput struct {
	Body string `json:"body"`
}

func (s *SocialService) hydrateGroup(g *domain.SocialGroup) {
	if g.Creator != nil {
		g.Creator.AvatarURL = s.profiles.UploadURL(g.Creator.AvatarURL)
	}
}

func (s *SocialService) hydrateGroupList(groups []domain.SocialGroup) []domain.SocialGroup {
	for i := range groups {
		s.hydrateGroup(&groups[i])
	}
	return groups
}

func (s *SocialService) CreateGroup(ctx context.Context, user *domain.User, input CreateSocialGroupInput) (*domain.SocialGroup, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 {
		return nil, errors.New("le nom doit faire entre 2 et 80 caractères")
	}
	desc := strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(desc) > 1000 {
		return nil, errors.New("description trop longue")
	}
	privacy := domain.SocialGroupPrivacy(strings.TrimSpace(input.Privacy))
	if privacy == "" {
		privacy = domain.SocialGroupApproval
	}
	if privacy != domain.SocialGroupOpen && privacy != domain.SocialGroupApproval {
		return nil, errors.New("privacy must be open or approval")
	}
	g := &domain.SocialGroup{
		Name:        name,
		Description: desc,
		Privacy:     privacy,
		CreatedBy:   user.ID,
	}
	if err := s.social.CreateGroup(ctx, g); err != nil {
		return nil, err
	}
	if err := s.social.AddGroupMember(ctx, g.ID, user.ID, domain.SocialGroupRoleAdmin); err != nil {
		return nil, err
	}
	return s.GetGroup(ctx, user, g.ID)
}

func (s *SocialService) GetGroup(ctx context.Context, viewer *domain.User, groupID uuid.UUID) (*domain.SocialGroup, error) {
	g, err := s.social.GetGroup(ctx, groupID, s.viewerID(viewer))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	s.hydrateGroup(g)
	return g, nil
}

func (s *SocialService) ListGroups(ctx context.Context, viewer *domain.User, mineOnly bool) ([]domain.SocialGroup, error) {
	if mineOnly {
		if err := s.requireUser(viewer); err != nil {
			return nil, err
		}
	}
	groups, err := s.social.ListGroups(ctx, s.viewerID(viewer), mineOnly, 40)
	if err != nil {
		return nil, err
	}
	return s.hydrateGroupList(groups), nil
}

func (s *SocialService) JoinGroup(ctx context.Context, user *domain.User, groupID uuid.UUID, input JoinSocialGroupInput) (*domain.SocialGroup, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	g, err := s.social.GetGroup(ctx, groupID, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	if g.JoinStatus == domain.SocialGroupJoinMember {
		return s.GetGroup(ctx, user, groupID)
	}
	msg := strings.TrimSpace(input.Message)
	if utf8.RuneCountInString(msg) > 500 {
		return nil, errors.New("message trop long")
	}
	if g.Privacy == domain.SocialGroupOpen {
		if err := s.social.AddGroupMember(ctx, groupID, user.ID, domain.SocialGroupRoleMember); err != nil {
			return nil, err
		}
		_ = s.social.TouchGroup(ctx, groupID)
	} else {
		if err := s.social.UpsertJoinRequest(ctx, groupID, user.ID, msg); err != nil {
			return nil, err
		}
	}
	return s.GetGroup(ctx, user, groupID)
}

func (s *SocialService) LeaveGroup(ctx context.Context, user *domain.User, groupID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	role, err := s.social.GetGroupMemberRole(ctx, groupID, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialGroupNotMember
	}
	if err != nil {
		return err
	}
	if role == domain.SocialGroupRoleAdmin {
		n, err := s.social.CountGroupAdmins(ctx, groupID)
		if err != nil {
			return err
		}
		if n <= 1 {
			return errors.New("nommez un autre admin avant de quitter")
		}
	}
	return s.social.RemoveGroupMember(ctx, groupID, user.ID)
}

func (s *SocialService) requireGroupAdmin(ctx context.Context, user *domain.User, groupID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	role, err := s.social.GetGroupMemberRole(ctx, groupID, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialGroupForbidden
	}
	if err != nil {
		return err
	}
	if role != domain.SocialGroupRoleAdmin {
		return ErrSocialGroupForbidden
	}
	return nil
}

func (s *SocialService) ListJoinRequests(ctx context.Context, user *domain.User, groupID uuid.UUID) ([]domain.SocialGroupJoinRequest, error) {
	if err := s.requireGroupAdmin(ctx, user, groupID); err != nil {
		return nil, err
	}
	reqs, err := s.social.ListJoinRequests(ctx, groupID)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].User != nil {
			reqs[i].User.AvatarURL = s.profiles.UploadURL(reqs[i].User.AvatarURL)
		}
	}
	return reqs, nil
}

func (s *SocialService) ApproveJoinRequest(ctx context.Context, user *domain.User, groupID, requesterID uuid.UUID) error {
	if err := s.requireGroupAdmin(ctx, user, groupID); err != nil {
		return err
	}
	err := s.social.ReviewJoinRequest(ctx, groupID, requesterID, user.ID, true)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialGroupNotFound
	}
	return err
}

func (s *SocialService) RejectJoinRequest(ctx context.Context, user *domain.User, groupID, requesterID uuid.UUID) error {
	if err := s.requireGroupAdmin(ctx, user, groupID); err != nil {
		return err
	}
	err := s.social.ReviewJoinRequest(ctx, groupID, requesterID, user.ID, false)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialGroupNotFound
	}
	return err
}

func (s *SocialService) ListGroupMembers(ctx context.Context, viewer *domain.User, groupID uuid.UUID) ([]domain.SocialGroupMember, error) {
	_, err := s.social.GetGroup(ctx, groupID, s.viewerID(viewer))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	members, err := s.social.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		if members[i].User != nil {
			members[i].User.AvatarURL = s.profiles.UploadURL(members[i].User.AvatarURL)
		}
	}
	return members, nil
}

func (s *SocialService) ListGroupMessages(ctx context.Context, user *domain.User, groupID uuid.UUID, limit int, before *time.Time) ([]domain.SocialGroupMessage, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	ok, err := s.social.IsGroupMember(ctx, groupID, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialGroupNotMember
	}
	msgs, err := s.social.ListGroupMessages(ctx, groupID, limit, before)
	if err != nil {
		return nil, err
	}
	for i := range msgs {
		if msgs[i].Sender != nil {
			msgs[i].Sender.AvatarURL = s.profiles.UploadURL(msgs[i].Sender.AvatarURL)
		}
	}
	return msgs, nil
}

func (s *SocialService) SendGroupMessage(ctx context.Context, user *domain.User, groupID uuid.UUID, input SendGroupMessageInput) (*domain.SocialGroupMessage, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	ok, err := s.social.IsGroupMember(ctx, groupID, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialGroupNotMember
	}
	body := strings.TrimSpace(input.Body)
	if body == "" || utf8.RuneCountInString(body) > 4000 {
		return nil, errors.New("message invalide")
	}
	msg := &domain.SocialGroupMessage{
		GroupID:  groupID,
		SenderID: user.ID,
		Body:     body,
	}
	if err := s.social.CreateGroupMessage(ctx, msg); err != nil {
		return nil, err
	}
	msg.Sender = &domain.SocialAuthor{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName,
		AvatarURL: s.profiles.UploadURL(user.AvatarPath),
	}
	return msg, nil
}
