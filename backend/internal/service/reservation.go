package service

import (
	"context"
	"fmt"
	"time"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type ReservationService struct {
	tx           repository.TxManager
	reservations repository.ReservationRepository
	items        repository.ToolItemRepository
	customers    repository.CustomerRepository
	status       *StatusService
	now          Clock
}

func NewReservationService(tx repository.TxManager, reservations repository.ReservationRepository,
	items repository.ToolItemRepository, customers repository.CustomerRepository,
	status *StatusService, now Clock) *ReservationService {
	return &ReservationService{tx: tx, reservations: reservations, items: items, customers: customers,
		status: status, now: now}
}

type CreateReservationInput struct {
	ToolTypeID *int64 // reserve any free item of this type
	ToolItemID *int64 // staff/admin only: reserve this exact item
	CustomerID *int64 // staff/admin only; customers always reserve for themselves
	StartDate  time.Time
	EndDate    time.Time
}

// startTolerance lets a reservation start "now" despite request latency.
const startTolerance = 5 * time.Minute

// Create books a tool item for [StartDate, EndDate). It locks the candidate
// item rows, so two concurrent bookings of the same item cannot both succeed.
// Customers name a tool type and the first free item is chosen; staff and admins
// may name an exact item. New reservations are CONFIRMED.
func (s *ReservationService) Create(ctx context.Context, p domain.Principal, in CreateReservationInput) (domain.Reservation, error) {
	var customerID int64
	if p.IsStaffOrAdmin() {
		if in.CustomerID == nil {
			return domain.Reservation{}, invalid("customer_id is required")
		}
		customerID = *in.CustomerID
	} else {
		if p.CustomerID == nil {
			return domain.Reservation{}, fmt.Errorf("%w: account has no customer record", domain.ErrForbidden)
		}
		if in.CustomerID != nil && *in.CustomerID != *p.CustomerID {
			return domain.Reservation{}, fmt.Errorf("%w: customers can only reserve for themselves", domain.ErrForbidden)
		}
		if in.ToolItemID != nil {
			return domain.Reservation{}, invalid("customers reserve by tool_type_id")
		}
		customerID = *p.CustomerID
	}
	if in.ToolTypeID == nil && in.ToolItemID == nil {
		return domain.Reservation{}, invalid("tool_type_id or tool_item_id is required")
	}
	if !in.EndDate.After(in.StartDate) {
		return domain.Reservation{}, invalid("end_date must be after start_date")
	}
	if in.StartDate.Before(s.now().Add(-startTolerance)) {
		return domain.Reservation{}, invalid("start_date cannot be in the past")
	}
	if err := s.status.Refresh(ctx); err != nil {
		return domain.Reservation{}, err
	}

	var created domain.Reservation
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.customers.GetByID(ctx, customerID); err != nil {
			return err
		}
		itemID, err := s.pickItem(ctx, in)
		if err != nil {
			return err
		}
		r := domain.Reservation{CustomerID: customerID, ToolItemID: itemID,
			StartDate: in.StartDate, EndDate: in.EndDate, Status: domain.ReservationConfirmed}
		if err := s.reservations.Create(ctx, &r); err != nil {
			return err
		}
		created, err = s.reservations.GetByID(ctx, r.ID)
		return err
	})
	return created, err
}

// pickItem locks and returns a free item for the period. Must run in a tx.
func (s *ReservationService) pickItem(ctx context.Context, in CreateReservationInput) (int64, error) {
	if in.ToolItemID != nil {
		it, err := s.items.GetByIDForUpdate(ctx, *in.ToolItemID)
		if err != nil {
			return 0, err
		}
		if in.ToolTypeID != nil && it.ToolTypeID != *in.ToolTypeID {
			return 0, invalid("tool item %d is not of tool type %d", it.ID, *in.ToolTypeID)
		}
		if !it.Status.Bookable() {
			return 0, fmt.Errorf("%w: item %d is %s", domain.ErrUnavailable, it.ID, it.Status)
		}
		busy, err := s.reservations.HasConflict(ctx, it.ID, in.StartDate, in.EndDate, 0)
		if err != nil {
			return 0, err
		}
		if busy {
			return 0, fmt.Errorf("%w: item %d is already booked in this period", domain.ErrUnavailable, it.ID)
		}
		return it.ID, nil
	}

	items, err := s.items.ListByTypeForUpdate(ctx, *in.ToolTypeID)
	if err != nil {
		return 0, err
	}
	for _, it := range items {
		if !it.Status.Bookable() {
			continue
		}
		busy, err := s.reservations.HasConflict(ctx, it.ID, in.StartDate, in.EndDate, 0)
		if err != nil {
			return 0, err
		}
		if !busy {
			return it.ID, nil
		}
	}
	return 0, fmt.Errorf("%w: no item of this type is free for the whole period", domain.ErrUnavailable)
}

type ReservationListFilter struct {
	CustomerID *int64
	ToolItemID *int64
	Statuses   []domain.ReservationStatus
	From, To   *time.Time
}

// List returns reservations; customers only see their own.
func (s *ReservationService) List(ctx context.Context, p domain.Principal, f ReservationListFilter) ([]domain.Reservation, error) {
	if !p.IsStaffOrAdmin() {
		if p.CustomerID == nil {
			return []domain.Reservation{}, nil
		}
		f.CustomerID = p.CustomerID
	}
	if err := s.status.Refresh(ctx); err != nil {
		return nil, err
	}
	return s.reservations.List(ctx, repository.ReservationFilter{
		CustomerID: f.CustomerID, ToolItemID: f.ToolItemID, Statuses: f.Statuses, From: f.From, To: f.To})
}

func (s *ReservationService) Get(ctx context.Context, p domain.Principal, id int64) (domain.Reservation, error) {
	if err := s.status.Refresh(ctx); err != nil {
		return domain.Reservation{}, err
	}
	r, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return r, err
	}
	if !ownsCustomerRecord(p, r.CustomerID) {
		return domain.Reservation{}, fmt.Errorf("reservation: %w", domain.ErrNotFound)
	}
	return r, nil
}

// Cancel cancels an open reservation. Customers can cancel only their own.
func (s *ReservationService) Cancel(ctx context.Context, p domain.Principal, id int64) (domain.Reservation, error) {
	if err := s.status.Refresh(ctx); err != nil {
		return domain.Reservation{}, err
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		r, err := s.reservations.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if !ownsCustomerRecord(p, r.CustomerID) {
			return fmt.Errorf("reservation: %w", domain.ErrNotFound)
		}
		if !r.Status.Open() {
			return fmt.Errorf("%w: reservation is %s", domain.ErrInvalidState, r.Status)
		}
		return s.reservations.UpdateStatus(ctx, id, domain.ReservationCancelled)
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	return s.reservations.GetByID(ctx, id)
}

// ownsCustomerRecord: staff and admins see everything, customers only their own.
func ownsCustomerRecord(p domain.Principal, customerID int64) bool {
	return p.IsStaffOrAdmin() || (p.CustomerID != nil && *p.CustomerID == customerID)
}
