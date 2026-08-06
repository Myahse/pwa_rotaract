package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/push"
	"github.com/rotaract-civ/backend/internal/repository"
)

type BirthdayService struct {
	birthdays    *repository.BirthdayRepository
	pushSubs     *repository.PushSubscriptionRepository
	profiles     *ProfileService
	push         *push.Sender
	location     *time.Location
	notifyHour   int
	appPublicURL string
	logger       *slog.Logger
}

func NewBirthdayService(
	birthdays *repository.BirthdayRepository,
	pushSubs *repository.PushSubscriptionRepository,
	profiles *ProfileService,
	pushSender *push.Sender,
	timezone string,
	notifyHour int,
	appPublicURL string,
	logger *slog.Logger,
) (*BirthdayService, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load birthday timezone: %w", err)
	}
	if notifyHour < 0 || notifyHour > 23 {
		notifyHour = 8
	}

	return &BirthdayService{
		birthdays:    birthdays,
		pushSubs:     pushSubs,
		profiles:     profiles,
		push:         pushSender,
		location:     loc,
		notifyHour:   notifyHour,
		appPublicURL: strings.TrimRight(appPublicURL, "/"),
		logger:       logger,
	}, nil
}

func (s *BirthdayService) VAPIDPublicKey() string {
	if s.push == nil {
		return ""
	}
	return s.push.PublicKey()
}

type PushSubscriptionInput struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

func (s *BirthdayService) SavePushSubscription(ctx context.Context, userID uuid.UUID, input PushSubscriptionInput, userAgent string) error {
	sub := &domain.PushSubscription{
		UserID:   userID,
		Endpoint: strings.TrimSpace(input.Endpoint),
		P256dh:   strings.TrimSpace(input.P256dh),
		Auth:     strings.TrimSpace(input.Auth),
	}
	if sub.Endpoint == "" || sub.P256dh == "" || sub.Auth == "" {
		return fmt.Errorf("endpoint, p256dh and auth are required")
	}
	if userAgent != "" {
		sub.UserAgent = &userAgent
	}
	return s.pushSubs.Upsert(ctx, sub)
}

func (s *BirthdayService) RemovePushSubscription(ctx context.Context, userID uuid.UUID, endpoint string) error {
	return s.pushSubs.DeleteByEndpoint(ctx, userID, endpoint)
}

func (s *BirthdayService) WidgetForUser(ctx context.Context, userID uuid.UUID) (*domain.BirthdayWidget, error) {
	month, day, _ := repository.TodayInLocation(s.location)
	now := time.Now().In(s.location)

	isMine, err := s.birthdays.UserHasBirthdayOn(ctx, userID, month, day)
	if err != nil {
		return nil, err
	}

	user, err := s.profiles.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	widget := &domain.BirthdayWidget{
		Date:         now.Format("2006-01-02"),
		IsMyBirthday: isMine,
		Emoji:        "🎂",
		Theme: domain.BirthdayWidgetTheme{
			Primary:  "#be034d",
			Accent:   "#f59e0b",
			Gradient: "linear-gradient(135deg, #be034d 0%, #d4356f 50%, #f59e0b 100%)",
		},
		UpdatedAt: now.UTC(),
	}

	if isMine {
		widget.Title = fmt.Sprintf("Joyeux anniversaire %s !", user.FirstName)
		widget.Message = fmt.Sprintf(
			"%s, toute la communauté Rotaract CIV te souhaite une journée remplie de joie, de réussite et de belles surprises.",
			user.FirstName,
		)
	} else {
		widget.Title = "Anniversaires du jour"
		widget.Message = "Pense à souhaiter un joyeux anniversaire à tes camarades Rotaract."
	}

	clubBirthdays, err := s.clubBirthdaysForUser(ctx, userID, month, day)
	if err != nil {
		return nil, err
	}
	widget.ClubBirthdays = clubBirthdays

	return widget, nil
}

func (s *BirthdayService) ClubBirthdaysToday(ctx context.Context, clubID uuid.UUID) ([]domain.BirthdayWidgetMember, error) {
	month, day, _ := repository.TodayInLocation(s.location)
	users, err := s.birthdays.ClubBirthdaysOn(ctx, clubID, month, day)
	if err != nil {
		return nil, err
	}

	items := make([]domain.BirthdayWidgetMember, 0, len(users))
	for i := range users {
		public := s.profiles.PublicUser(&users[i])
		items = append(items, domain.BirthdayWidgetMember{
			UserID:    public.ID,
			FirstName: public.FirstName,
			LastName:  public.LastName,
			AvatarURL: public.AvatarURL,
			Message:   fmt.Sprintf("C'est l'anniversaire de %s aujourd'hui !", public.FirstName),
		})
	}
	return items, nil
}

func (s *BirthdayService) RunDailyNotifications(ctx context.Context) (int, error) {
	month, day, year := repository.TodayInLocation(s.location)
	users, err := s.birthdays.UsersWithBirthdayOn(ctx, month, day)
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, item := range users {
		already, err := s.birthdays.WasNotified(ctx, item.User.ID, year)
		if err != nil {
			s.logger.Error("birthday notify check failed", "user_id", item.User.ID, "error", err)
			continue
		}
		if already {
			continue
		}

		if err := s.sendBirthdayPush(ctx, item, year); err != nil {
			s.logger.Error("birthday push failed", "user_id", item.User.ID, "error", err)
			continue
		}

		if err := s.birthdays.MarkNotified(ctx, item.User.ID, year); err != nil {
			s.logger.Error("birthday mark notified failed", "user_id", item.User.ID, "error", err)
			continue
		}
		sent++
	}

	if sent > 0 {
		s.logger.Info("birthday notifications sent", "count", sent, "date", fmt.Sprintf("%02d-%02d", month, day))
	}
	return sent, nil
}

func (s *BirthdayService) sendBirthdayPush(ctx context.Context, item repository.BirthdayUser, year int) error {
	if !s.push.Enabled() {
		s.logger.Info("birthday push skipped, VAPID not configured", "user_id", item.User.ID)
		return nil
	}

	subs, err := s.pushSubs.ListByUser(ctx, item.User.ID)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}

	clubLine := "Toute la communauté Rotaract CIV"
	if item.ClubName != nil && *item.ClubName != "" {
		clubLine = fmt.Sprintf("Toute l'équipe %s", *item.ClubName)
	}

	title := fmt.Sprintf("Joyeux anniversaire %s ! 🎉", item.User.FirstName)
	body := fmt.Sprintf("%s te souhaite une excellente journée.", clubLine)
	tag := fmt.Sprintf("birthday-%d", year)

	var lastErr error
	for _, sub := range subs {
		err := s.push.Send(push.Subscription{
			Endpoint: sub.Endpoint,
			P256dh:   sub.P256dh,
			Auth:     sub.Auth,
		}, push.NotificationPayload{
			Title: title,
			Body:  body,
			Icon:  s.appPublicURL + "/icons/birthday-192.png",
			Badge: s.appPublicURL + "/icons/badge-72.png",
			Tag:   tag,
			Data: map[string]any{
				"type": "birthday",
				"url":  "/birthday",
			},
		})
		if err != nil {
			lastErr = err
			_ = s.pushSubs.DeleteByEndpointGlobal(ctx, sub.Endpoint)
		}
	}
	return lastErr
}

func (s *BirthdayService) clubBirthdaysForUser(ctx context.Context, userID uuid.UUID, month, day int) ([]domain.BirthdayWidgetMember, error) {
	rows, err := s.birthdays.ClubmatesBirthdaysOn(ctx, userID, month, day)
	if err != nil {
		return nil, err
	}

	items := make([]domain.BirthdayWidgetMember, 0, len(rows))
	for _, item := range rows {
		public := s.profiles.PublicUser(&item.User)
		items = append(items, domain.BirthdayWidgetMember{
			UserID:    public.ID,
			FirstName: public.FirstName,
			LastName:  public.LastName,
			AvatarURL: public.AvatarURL,
			Message:   fmt.Sprintf("C'est l'anniversaire de %s aujourd'hui !", public.FirstName),
		})
	}
	return items, nil
}

func (s *BirthdayService) Location() *time.Location {
	return s.location
}

func (s *BirthdayService) NotifyHour() int {
	return s.notifyHour
}
