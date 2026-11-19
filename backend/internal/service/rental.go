package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type RentalService struct {
	tx           repository.TxManager
	rentals      repository.RentalRepository
	reservations repository.ReservationRepository
	items        repository.ToolItemRepository
	types        repository.ToolTypeRepository
	status       *StatusService
	now          Clock
}

func NewRentalService(tx repository.TxManager, rentals repository.RentalRepository,
	reservations repository.ReservationRepository, items repository.ToolItemRepository,
	types repository.ToolTypeRepository, status *StatusService, now Clock) *RentalService {
	return &RentalService{tx: tx, rentals: rentals, reservations: reservations, items: items, types: types,
		status: status, now: now}
}

type HandoverInput struct {
	ReservationID int64
	// Condition at handover; defaults to the item's current condition.
	Condition *domain.Condition
}

// Handover starts a rental from an open reservation. The rental starts now and
// is expected to end at the reservation's end. The deposit is the tool type's
// deposit. Items in maintenance, broken, retired or still rented out are refused.
func (s *RentalService) Handover(ctx context.Context, p domain.Principal, in HandoverInput) (domain.Rental, error) {
	if err := s.status.Refresh(ctx); err != nil {
		return domain.Rental{}, err
	}
	var id int64
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		now := s.now()
		res, err := s.reservations.GetByIDForUpdate(ctx, in.ReservationID)
		if err != nil {
			return err
		}
		if !res.Status.Open() {
			return fmt.Errorf("%w: reservation is %s", domain.ErrInvalidState, res.Status)
		}
		if !res.EndDate.After(now) {
			return fmt.Errorf("%w: reservation period is over", domain.ErrInvalidState)
		}
		it, err := s.items.GetByIDForUpdate(ctx, res.ToolItemID)
		if err != nil {
			return err
		}
		if it.Status != domain.ToolAvailable {
			return fmt.Errorf("%w: item %d is %s", domain.ErrUnavailable, it.ID, it.Status)
		}
		busy, err := s.reservations.HasConflict(ctx, it.ID, now, res.EndDate, res.ID)
		if err != nil {
			return err
		}
		if busy {
			return fmt.Errorf("%w: item %d is booked by someone else before the reservation starts",
				domain.ErrUnavailable, it.ID)
		}
		cond := it.CurrentCondition
		if in.Condition != nil {
			cond = *in.Condition
		}
		if !cond.Valid() {
			return invalid("unknown condition %q", cond)
		}
		if cond == domain.ConditionDamaged {
			return fmt.Errorf("%w: a DAMAGED item cannot be handed over", domain.ErrUnavailable)
		}
		tt, err := s.types.GetByID(ctx, it.ToolTypeID)
		if err != nil {
			return err
		}
		r := domain.Rental{
			CustomerID: res.CustomerID, ToolItemID: it.ID, ReservationID: res.ID, StaffID: p.UserID,
			StartDate: now, ExpectedEndDate: res.EndDate, DepositPaid: tt.Deposit,
			ConditionAtHandover: cond, Status: domain.RentalActive,
		}
		if err := s.rentals.Create(ctx, &r); err != nil {
			return err
		}
		if err := s.reservations.UpdateStatus(ctx, res.ID, domain.ReservationConverted); err != nil {
			return err
		}
		if err := s.items.UpdateState(ctx, it.ID, domain.ToolRented, cond); err != nil {
			return err
		}
		id = r.ID
		return nil
	})
	if err != nil {
		return domain.Rental{}, err
	}
	return s.rentals.GetByID(ctx, id)
}

type ReturnInput struct {
	Condition      domain.Condition
	OperatingHours *float64   // hours the tool ran, for maintenance tracking
	ReturnedAt     *time.Time // defaults to now
}

// ReturnResult is the closed rental plus the fee breakdown.
type ReturnResult struct {
	Rental domain.Rental `json:"rental"`
	Fees   domain.Fees   `json:"fees"`
}

// Return closes a rental and computes the fees (see domain.CalculateFees).
// The item becomes AVAILABLE again, or BROKEN if it came back DAMAGED.
func (s *RentalService) Return(ctx context.Context, id int64, in ReturnInput) (ReturnResult, error) {
	if !in.Condition.Valid() {
		return ReturnResult{}, invalid("condition is required (NEW, GOOD, FAIR, POOR or DAMAGED)")
	}
	if in.OperatingHours != nil && *in.OperatingHours < 0 {
		return ReturnResult{}, invalid("operating_hours cannot be negative")
	}
	var fees domain.Fees
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		now := s.now()
		r, err := s.rentals.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if !r.Status.Open() {
			return fmt.Errorf("%w: rental is %s", domain.ErrInvalidState, r.Status)
		}
		returnedAt := now
		if in.ReturnedAt != nil {
			returnedAt = *in.ReturnedAt
		}
		if returnedAt.Before(r.StartDate) || returnedAt.After(now.Add(startTolerance)) {
			return invalid("returned_at must be between the rental start and now")
		}
		it, err := s.items.GetByIDForUpdate(ctx, r.ToolItemID)
		if err != nil {
			return err
		}
		tt, err := s.types.GetByID(ctx, it.ToolTypeID)
		if err != nil {
			return err
		}
		fees = domain.CalculateFees(tt.DailyRate, r.StartDate, r.ExpectedEndDate, returnedAt)
		cond := in.Condition
		r.ActualEndDate = &returnedAt
		r.ConditionAtReturn = &cond
		r.BaseFee, r.LateFee, r.TotalFee = &fees.BaseFee, &fees.LateFee, &fees.TotalFee
		r.OperatingHours = in.OperatingHours
		r.Status = domain.RentalReturned
		if err := s.rentals.Close(ctx, &r); err != nil {
			return err
		}
		return s.items.UpdateState(ctx, it.ID, itemStatusAfterReturn(cond), cond)
	})
	if err != nil {
		return ReturnResult{}, err
	}
	r, err := s.rentals.GetByID(ctx, id)
	return ReturnResult{Rental: r, Fees: fees}, err
}

type CancelRentalInput struct {
	Condition *domain.Condition // defaults to the condition at handover
}

// Cancel closes a rental free of charge, when the tool comes back unused.
func (s *RentalService) Cancel(ctx context.Context, id int64, in CancelRentalInput) (domain.Rental, error) {
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		now := s.now()
		r, err := s.rentals.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if !r.Status.Open() {
			return fmt.Errorf("%w: rental is %s", domain.ErrInvalidState, r.Status)
		}
		cond := r.ConditionAtHandover
		if in.Condition != nil {
			cond = *in.Condition
		}
		if !cond.Valid() {
			return invalid("unknown condition %q", cond)
		}
		zero := decimal.Zero
		r.ActualEndDate = &now
		r.ConditionAtReturn = &cond
		r.BaseFee, r.LateFee, r.TotalFee = &zero, &zero, &zero
		r.Status = domain.RentalCancelled
		if err := s.rentals.Close(ctx, &r); err != nil {
			return err
		}
		if _, err := s.items.GetByIDForUpdate(ctx, r.ToolItemID); err != nil {
			return err
		}
		return s.items.UpdateState(ctx, r.ToolItemID, itemStatusAfterReturn(cond), cond)
	})
	if err != nil {
		return domain.Rental{}, err
	}
	return s.rentals.GetByID(ctx, id)
}

func itemStatusAfterReturn(c domain.Condition) domain.ToolStatus {
	if c == domain.ConditionDamaged {
		return domain.ToolBroken
	}
	return domain.ToolAvailable
}

type RentalListFilter struct {
	CustomerID *int64
	ToolItemID *int64
	StaffID    *int64
	Statuses   []domain.RentalStatus
}

// List returns rentals; customers only see their own.
func (s *RentalService) List(ctx context.Context, p domain.Principal, f RentalListFilter) ([]domain.Rental, error) {
	if !p.IsStaffOrAdmin() {
		if p.CustomerID == nil {
			return []domain.Rental{}, nil
		}
		f.CustomerID = p.CustomerID
	}
	if err := s.status.Refresh(ctx); err != nil {
		return nil, err
	}
	return s.rentals.List(ctx, repository.RentalFilter{
		CustomerID: f.CustomerID, ToolItemID: f.ToolItemID, StaffID: f.StaffID, Statuses: f.Statuses})
}

func (s *RentalService) Get(ctx context.Context, p domain.Principal, id int64) (domain.Rental, error) {
	if err := s.status.Refresh(ctx); err != nil {
		return domain.Rental{}, err
	}
	r, err := s.rentals.GetByID(ctx, id)
	if err != nil {
		return r, err
	}
	if !ownsCustomerRecord(p, r.CustomerID) {
		return domain.Rental{}, fmt.Errorf("rental: %w", domain.ErrNotFound)
	}
	return r, nil
}
