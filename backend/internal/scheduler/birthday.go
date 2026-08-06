package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/rotaract-civ/backend/internal/service"
)

type BirthdayScheduler struct {
	birthdays *service.BirthdayService
	logger    *slog.Logger
	mu        sync.Mutex
	lastRun   string
}

func NewBirthdayScheduler(birthdays *service.BirthdayService, logger *slog.Logger) *BirthdayScheduler {
	return &BirthdayScheduler{birthdays: birthdays, logger: logger}
}

func (s *BirthdayScheduler) Run(ctx context.Context) {
	loc := s.birthdays.Location()
	hour := s.birthdays.NotifyHour()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	s.logger.Info("birthday scheduler started", "timezone", loc.String(), "notify_hour", hour)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("birthday scheduler stopped")
			return
		case now := <-ticker.C:
			s.maybeRun(ctx, now.In(loc), hour)
		}
	}
}

func (s *BirthdayScheduler) maybeRun(ctx context.Context, now time.Time, hour int) {
	if now.Hour() != hour || now.Minute() != 0 {
		return
	}

	dateKey := now.Format("2006-01-02")
	s.mu.Lock()
	if s.lastRun == dateKey {
		s.mu.Unlock()
		return
	}
	s.lastRun = dateKey
	s.mu.Unlock()

	if _, err := s.birthdays.RunDailyNotifications(ctx); err != nil {
		s.logger.Error("daily birthday job failed", "error", err)
	}
}

func (s *BirthdayScheduler) RunNow(ctx context.Context) (int, error) {
	return s.birthdays.RunDailyNotifications(ctx)
}
