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
	ErrSocialUserNotFound   = errors.New("user not found")
	ErrSocialNotFriends     = errors.New("friends only")
	ErrSocialCommentNotFound = errors.New("comment not found")
	ErrSocialConversationNF = errors.New("conversation not found")
)

type CreateSocialCommentInput struct {
	Body     string     `json:"body"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type SendSocialMessageInput struct {
	Body            string     `json:"body"`
	SharedCommentID *uuid.UUID `json:"shared_comment_id"`
	SharedPostID    *uuid.UUID `json:"shared_post_id"`
}

type ShareCommentInput struct {
	FriendUserID uuid.UUID `json:"friend_user_id"`
	Note         string    `json:"note"`
}

func (s *SocialService) CreateComment(ctx context.Context, user *domain.User, postID uuid.UUID, input CreateSocialCommentInput) (*domain.SocialComment, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	ok, err := s.social.PostExistsVisible(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialPostNotFound
	}
	body := strings.TrimSpace(input.Body)
	if body == "" || utf8.RuneCountInString(body) > 2000 {
		return nil, errors.New("comment must be between 1 and 2000 characters")
	}

	var parentID *uuid.UUID
	if input.ParentID != nil {
		parent, err := s.social.GetComment(ctx, *input.ParentID)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSocialCommentNotFound
		}
		if err != nil {
			return nil, err
		}
		if parent.PostID != postID {
			return nil, errors.New("reply must belong to the same post")
		}
		if parent.ParentID != nil {
			return nil, errors.New("only one reply level is allowed")
		}
		parentID = input.ParentID
	}

	comment := &domain.SocialComment{
		PostID:   postID,
		AuthorID: user.ID,
		ParentID: parentID,
		Body:     body,
	}
	if err := s.social.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	comment.Author = &domain.SocialAuthor{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: s.profiles.UploadURL(user.AvatarPath),
	}
	comment.Replies = []domain.SocialComment{}
	return comment, nil
}

func (s *SocialService) GetUserProfile(ctx context.Context, viewer *domain.User, userID uuid.UUID) (*domain.SocialUserProfile, error) {
	profile, err := s.social.GetUserProfile(ctx, userID, s.viewerID(viewer))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialUserNotFound
	}
	if err != nil {
		return nil, err
	}
	profile.AvatarURL = s.profiles.UploadURL(profile.AvatarURL)
	clubs, err := s.social.ListUserClubs(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.Clubs = clubs
	return profile, nil
}

func (s *SocialService) ListUserPosts(ctx context.Context, viewer *domain.User, userID uuid.UUID, limit int) ([]domain.SocialPost, error) {
	ok, err := s.social.UserExistsActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialUserNotFound
	}
	rows, err := s.social.ListUserPosts(ctx, userID, s.viewerID(viewer), limit)
	if err != nil {
		return nil, err
	}
	return s.attachMedia(ctx, rows)
}

func (s *SocialService) RequestFriend(ctx context.Context, user *domain.User, targetID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	if user.ID == targetID {
		return errors.New("cannot friend yourself")
	}
	ok, err := s.social.UserExistsActive(ctx, targetID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSocialUserNotFound
	}
	return s.social.RequestFriendship(ctx, user.ID, targetID)
}

func (s *SocialService) AcceptFriend(ctx context.Context, user *domain.User, requesterID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	err := s.social.RespondFriendship(ctx, user.ID, requesterID, true)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialUserNotFound
	}
	return err
}

func (s *SocialService) DeclineFriend(ctx context.Context, user *domain.User, requesterID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	err := s.social.RespondFriendship(ctx, user.ID, requesterID, false)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialUserNotFound
	}
	return err
}

func (s *SocialService) RemoveFriend(ctx context.Context, user *domain.User, otherID uuid.UUID) error {
	if err := s.requireUser(user); err != nil {
		return err
	}
	err := s.social.RemoveFriendship(ctx, user.ID, otherID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSocialUserNotFound
	}
	return err
}

func (s *SocialService) ListFriends(ctx context.Context, user *domain.User) ([]domain.SocialFriendProfile, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	friends, err := s.social.ListFriends(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for i := range friends {
		friends[i].AvatarURL = s.profiles.UploadURL(friends[i].AvatarURL)
	}
	return friends, nil
}

func (s *SocialService) ListFriendRequests(ctx context.Context, user *domain.User) ([]domain.SocialFriendProfile, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	reqs, err := s.social.ListFriendRequests(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		reqs[i].AvatarURL = s.profiles.UploadURL(reqs[i].AvatarURL)
	}
	return reqs, nil
}

func (s *SocialService) ListConversations(ctx context.Context, user *domain.User) ([]domain.SocialConversation, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	convs, err := s.social.ListConversations(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for i := range convs {
		if convs[i].Peer != nil {
			convs[i].Peer.AvatarURL = s.profiles.UploadURL(convs[i].Peer.AvatarURL)
		}
	}
	return convs, nil
}

func (s *SocialService) OpenConversation(ctx context.Context, user *domain.User, peerID uuid.UUID) (*domain.SocialConversation, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	if user.ID == peerID {
		return nil, errors.New("cannot message yourself")
	}
	friends, err := s.social.AreFriends(ctx, user.ID, peerID)
	if err != nil {
		return nil, err
	}
	if !friends {
		return nil, ErrSocialNotFriends
	}
	id, err := s.social.GetOrCreateDM(ctx, user.ID, peerID)
	if err != nil {
		return nil, err
	}
	convs, err := s.social.ListConversations(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for i := range convs {
		if convs[i].ID == id {
			if convs[i].Peer != nil {
				convs[i].Peer.AvatarURL = s.profiles.UploadURL(convs[i].Peer.AvatarURL)
			}
			return &convs[i], nil
		}
	}
	return &domain.SocialConversation{ID: id, UpdatedAt: time.Now()}, nil
}

func (s *SocialService) ListMessages(ctx context.Context, user *domain.User, conversationID uuid.UUID, limit int, before *time.Time) ([]domain.SocialMessage, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	ok, err := s.social.IsConversationMember(ctx, conversationID, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialConversationNF
	}
	msgs, err := s.social.ListMessages(ctx, conversationID, limit, before)
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

func (s *SocialService) SendMessage(ctx context.Context, user *domain.User, conversationID uuid.UUID, input SendSocialMessageInput) (*domain.SocialMessage, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	ok, err := s.social.IsConversationMember(ctx, conversationID, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSocialConversationNF
	}
	body := strings.TrimSpace(input.Body)
	if body == "" && input.SharedCommentID == nil && input.SharedPostID == nil {
		return nil, errors.New("message cannot be empty")
	}
	if utf8.RuneCountInString(body) > 4000 {
		return nil, errors.New("message too long")
	}
	if input.SharedCommentID != nil {
		if _, err := s.social.GetComment(ctx, *input.SharedCommentID); errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSocialCommentNotFound
		} else if err != nil {
			return nil, err
		}
	}
	msg := &domain.SocialMessage{
		ConversationID:  conversationID,
		SenderID:        user.ID,
		Body:            body,
		SharedCommentID: input.SharedCommentID,
		SharedPostID:    input.SharedPostID,
	}
	if err := s.social.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	msg.Sender = &domain.SocialAuthor{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName,
		AvatarURL: s.profiles.UploadURL(user.AvatarPath),
	}
	if msg.SharedCommentID != nil {
		c, err := s.social.GetComment(ctx, *msg.SharedCommentID)
		if err == nil {
			if c.Author != nil {
				c.Author.AvatarURL = s.profiles.UploadURL(c.Author.AvatarURL)
			}
			msg.SharedComment = c
		}
	}
	return msg, nil
}

func (s *SocialService) ShareCommentWithFriend(ctx context.Context, user *domain.User, commentID uuid.UUID, input ShareCommentInput) (*domain.SocialMessage, error) {
	if err := s.requireUser(user); err != nil {
		return nil, err
	}
	comment, err := s.social.GetComment(ctx, commentID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSocialCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	friends, err := s.social.AreFriends(ctx, user.ID, input.FriendUserID)
	if err != nil {
		return nil, err
	}
	if !friends {
		return nil, ErrSocialNotFriends
	}
	convID, err := s.social.GetOrCreateDM(ctx, user.ID, input.FriendUserID)
	if err != nil {
		return nil, err
	}
	note := strings.TrimSpace(input.Note)
	if note == "" {
		note = "Regarde ce commentaire"
	}
	return s.SendMessage(ctx, user, convID, SendSocialMessageInput{
		Body:            note,
		SharedCommentID: &comment.ID,
		SharedPostID:    &comment.PostID,
	})
}
