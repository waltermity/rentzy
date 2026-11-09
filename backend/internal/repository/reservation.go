package repository

import (
	"context"
	"time"

	"rentzy/backend/internal/domain"
)

type ReservationFilter struct {
	CustomerID *int64
	ToolItemID *int64
	Statuses   []domain.ReservationStatus
	From, To   *time.Time // overlapping [From, To)
}

type ReservationRepository interface {
	List(ctx context.Context, f ReservationFilter) ([]domain.Reservation, error)
	GetByID(ctx context.Context, id int64) (domain.Reservation, error)
	GetByIDForUpdate(ctx context.Context, id int64) (domain.Reservation, error)
	Create(ctx context.Context, r *domain.Reservation) error
	UpdateStatus(ctx context.Context, id int64, status domain.ReservationStatus) error

	// HasConflict reports whether [start, end) overlaps an open reservation or an
	// open rental of the item. excludeReservationID (0 = none) is ignored.
	HasConflict(ctx context.Context, toolItemID int64, start, end time.Time, excludeReservationID int64) (bool, error)
	// BusyIntervals lists the open reservations and rentals of the items of a type
	// that overlap [from, to).
	BusyIntervals(ctx context.Context, toolTypeID int64, from, to time.Time) ([]domain.BusyInterval, error)

	// ExpireStale sets open reservations whose start is before cutoff to EXPIRED.
	ExpireStale(ctx context.Context, cutoff time.Time) (int64, error)
}
