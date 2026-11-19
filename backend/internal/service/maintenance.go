package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type MaintenanceService struct {
	tx          repository.TxManager
	maintenance repository.MaintenanceRepository
	items       repository.ToolItemRepository
	now         Clock
}

func NewMaintenanceService(tx repository.TxManager, maintenance repository.MaintenanceRepository,
	items repository.ToolItemRepository, now Clock) *MaintenanceService {
	return &MaintenanceService{tx: tx, maintenance: maintenance, items: items, now: now}
}

type StartMaintenanceInput struct {
	Type        domain.MaintenanceType
	Description string
	PerformedBy string
	StartedAt   *time.Time // defaults to now
}

// Start opens a maintenance record and puts the item into MAINTENANCE, so it
// can be neither reserved nor rented until the work is completed.
func (s *MaintenanceService) Start(ctx context.Context, itemID int64, in StartMaintenanceInput) (domain.Maintenance, error) {
	if !in.Type.Valid() {
		return domain.Maintenance{}, invalid("type must be SCHEDULED, REPAIR or INSPECTION")
	}
	started := s.now()
	if in.StartedAt != nil {
		started = *in.StartedAt
	}
	var id int64
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		it, err := s.items.GetByIDForUpdate(ctx, itemID)
		if err != nil {
			return err
		}
		switch it.Status {
		case domain.ToolRented:
			return fmt.Errorf("%w: item is rented out; record the return first", domain.ErrInvalidState)
		case domain.ToolRetired:
			return fmt.Errorf("%w: item is retired", domain.ErrInvalidState)
		}
		m := domain.Maintenance{ToolItemID: itemID, Type: in.Type, StartedAt: started, Cost: decimal.Zero,
			Description: strings.TrimSpace(in.Description), PerformedBy: strings.TrimSpace(in.PerformedBy)}
		if err := s.maintenance.Create(ctx, &m); err != nil {
			return err
		}
		id = m.ID
		return s.items.UpdateState(ctx, itemID, domain.ToolMaintenance, it.CurrentCondition)
	})
	if err != nil {
		return domain.Maintenance{}, err
	}
	return s.maintenance.GetByID(ctx, id)
}

type CompleteMaintenanceInput struct {
	Cost        decimal.Decimal
	Description *string // replaces the description when set
	PerformedBy *string
	CompletedAt *time.Time // defaults to now
	// Item status after the work: AVAILABLE (default), BROKEN or RETIRED.
	ResultStatus domain.ToolStatus
	// Item condition after the work; defaults to the current one.
	Condition *domain.Condition
}

// Complete closes a maintenance record and sets the item's status and
// condition. If the item has other open records it stays in MAINTENANCE.
func (s *MaintenanceService) Complete(ctx context.Context, id int64, in CompleteMaintenanceInput) (domain.Maintenance, error) {
	if in.Cost.IsNegative() {
		return domain.Maintenance{}, invalid("cost cannot be negative")
	}
	if in.ResultStatus == "" {
		in.ResultStatus = domain.ToolAvailable
	}
	switch in.ResultStatus {
	case domain.ToolAvailable, domain.ToolBroken, domain.ToolRetired:
	default:
		return domain.Maintenance{}, invalid("result_status must be AVAILABLE, BROKEN or RETIRED")
	}
	if in.Condition != nil && !in.Condition.Valid() {
		return domain.Maintenance{}, invalid("unknown condition %q", *in.Condition)
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.maintenance.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if m.CompletedAt != nil {
			return fmt.Errorf("%w: maintenance is already completed", domain.ErrInvalidState)
		}
		it, err := s.items.GetByIDForUpdate(ctx, m.ToolItemID)
		if err != nil {
			return err
		}
		completed := s.now()
		if in.CompletedAt != nil {
			completed = *in.CompletedAt
		}
		if completed.Before(m.StartedAt) {
			return invalid("completed_at cannot be before started_at")
		}
		m.CompletedAt = &completed
		m.Cost = in.Cost
		if in.Description != nil {
			m.Description = strings.TrimSpace(*in.Description)
		}
		if in.PerformedBy != nil {
			m.PerformedBy = strings.TrimSpace(*in.PerformedBy)
		}
		if err := s.maintenance.Complete(ctx, &m); err != nil {
			return err
		}

		cond := it.CurrentCondition
		if in.Condition != nil {
			cond = *in.Condition
		}
		open, err := s.maintenance.CountOpen(ctx, it.ID)
		if err != nil {
			return err
		}
		status := in.ResultStatus
		if open > 0 && status != domain.ToolRetired {
			status = domain.ToolMaintenance // other work still in progress
		}
		if status == domain.ToolAvailable && cond == domain.ConditionDamaged {
			return invalid("a DAMAGED item cannot become AVAILABLE; set condition or result_status")
		}
		return s.items.UpdateState(ctx, it.ID, status, cond)
	})
	if err != nil {
		return domain.Maintenance{}, err
	}
	return s.maintenance.GetByID(ctx, id)
}

func (s *MaintenanceService) List(ctx context.Context, f repository.MaintenanceFilter) ([]domain.Maintenance, error) {
	return s.maintenance.List(ctx, f)
}

func (s *MaintenanceService) Get(ctx context.Context, id int64) (domain.Maintenance, error) {
	return s.maintenance.GetByID(ctx, id)
}

// Due checks every non-retired item against its type's thresholds (rentals,
// days, operating hours since the last completed maintenance). onlyDue drops
// the items that are not due.
func (s *MaintenanceService) Due(ctx context.Context, onlyDue bool) ([]domain.MaintenanceDue, error) {
	usage, err := s.maintenance.Usage(ctx, nil)
	if err != nil {
		return nil, err
	}
	now := s.now()
	out := []domain.MaintenanceDue{}
	for _, u := range usage {
		d := domain.CheckDue(u, now)
		if onlyDue && !d.Due {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}
