package repository

import (
	"context"
	"time"

	"rentzy/backend/internal/domain"
)

type RentalFilter struct {
	CustomerID *int64
	ToolItemID *int64
	StaffID    *int64
	Statuses   []domain.RentalStatus
}

type RentalRepository interface {
	List(ctx context.Context, f RentalFilter) ([]domain.Rental, error)
	GetByID(ctx context.Context, id int64) (domain.Rental, error)
	GetByIDForUpdate(ctx context.Context, id int64) (domain.Rental, error)
	Create(ctx context.Context, r *domain.Rental) error
	// Close stores the return/cancel data: actual end, return condition, fees,
	// operating hours and status.
	Close(ctx context.Context, r *domain.Rental) error
	// MarkOverdue sets ACTIVE rentals whose expected end is before now to OVERDUE.
	MarkOverdue(ctx context.Context, now time.Time) (int64, error)
}
