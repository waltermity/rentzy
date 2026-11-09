package service

import (
	"context"
	"log"
	"time"

	"rentzy/backend/internal/repository"
)

// StatusService moves time-driven statuses forward:
//   - ACTIVE rentals past their expected end become OVERDUE;
//   - open reservations not handed over within the grace period after their
//     start become EXPIRED.
//
// It runs periodically in the background (Run), and services also call Refresh
// before reads and bookings so they never see stale statuses.
type StatusService struct {
	rentals      repository.RentalRepository
	reservations repository.ReservationRepository
	grace        time.Duration
	now          Clock
}

func NewStatusService(rentals repository.RentalRepository, reservations repository.ReservationRepository,
	grace time.Duration, now Clock) *StatusService {
	return &StatusService{rentals: rentals, reservations: reservations, grace: grace, now: now}
}

func (s *StatusService) Refresh(ctx context.Context) error {
	now := s.now()
	if _, err := s.rentals.MarkOverdue(ctx, now); err != nil {
		return err
	}
	_, err := s.reservations.ExpireStale(ctx, now.Add(-s.grace))
	return err
}

// Run calls Refresh every interval until ctx is cancelled.
func (s *StatusService) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if err := s.Refresh(ctx); err != nil && ctx.Err() == nil {
			log.Printf("status refresh: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
