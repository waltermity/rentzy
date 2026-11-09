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

type ToolService struct {
	tx           repository.TxManager
	categories   repository.CategoryRepository
	types        repository.ToolTypeRepository
	items        repository.ToolItemRepository
	reservations repository.ReservationRepository
	maintenance  repository.MaintenanceRepository
	status       *StatusService
	now          Clock
}

func NewToolService(tx repository.TxManager, categories repository.CategoryRepository, types repository.ToolTypeRepository,
	items repository.ToolItemRepository, reservations repository.ReservationRepository,
	maintenance repository.MaintenanceRepository, status *StatusService, now Clock) *ToolService {
	return &ToolService{tx: tx, categories: categories, types: types, items: items,
		reservations: reservations, maintenance: maintenance, status: status, now: now}
}

// ---------- categories ----------

func (s *ToolService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

func (s *ToolService) CreateCategory(ctx context.Context, name string) (domain.Category, error) {
	name, err := requireText("name", name)
	if err != nil {
		return domain.Category{}, err
	}
	c := domain.Category{Name: name}
	return c, s.categories.Create(ctx, &c)
}

func (s *ToolService) UpdateCategory(ctx context.Context, id int64, name string) (domain.Category, error) {
	name, err := requireText("name", name)
	if err != nil {
		return domain.Category{}, err
	}
	c := domain.Category{ID: id, Name: name}
	return c, s.categories.Update(ctx, &c)
}

// ---------- tool types ----------

type ToolTypeInput struct {
	Name                    string
	CategoryID              int64
	Manufacturer            string
	Model                   string
	PurchasePrice           decimal.Decimal
	DailyRate               decimal.Decimal
	Deposit                 decimal.Decimal
	MaintenanceEveryRentals *int
	MaintenanceEveryDays    *int
	MaintenanceEveryHours   *float64
}

func (in ToolTypeInput) toDomain() (domain.ToolType, error) {
	name, err := requireText("name", in.Name)
	if err != nil {
		return domain.ToolType{}, err
	}
	if in.CategoryID <= 0 {
		return domain.ToolType{}, invalid("category_id is required")
	}
	if in.PurchasePrice.IsNegative() || in.Deposit.IsNegative() {
		return domain.ToolType{}, invalid("purchase_price and deposit cannot be negative")
	}
	if !in.DailyRate.IsPositive() {
		return domain.ToolType{}, invalid("daily_rate must be positive")
	}
	if (in.MaintenanceEveryRentals != nil && *in.MaintenanceEveryRentals <= 0) ||
		(in.MaintenanceEveryDays != nil && *in.MaintenanceEveryDays <= 0) ||
		(in.MaintenanceEveryHours != nil && *in.MaintenanceEveryHours <= 0) {
		return domain.ToolType{}, invalid("maintenance thresholds must be positive when set")
	}
	return domain.ToolType{
		Name: name, CategoryID: in.CategoryID,
		Manufacturer: strings.TrimSpace(in.Manufacturer), Model: strings.TrimSpace(in.Model),
		PurchasePrice: in.PurchasePrice, DailyRate: in.DailyRate, Deposit: in.Deposit,
		MaintenanceEveryRentals: in.MaintenanceEveryRentals,
		MaintenanceEveryDays:    in.MaintenanceEveryDays,
		MaintenanceEveryHours:   in.MaintenanceEveryHours,
	}, nil
}

func (s *ToolService) ListTypes(ctx context.Context, f repository.ToolTypeFilter) ([]domain.ToolType, error) {
	return s.types.List(ctx, f)
}

func (s *ToolService) GetType(ctx context.Context, id int64) (domain.ToolType, error) {
	return s.types.GetByID(ctx, id)
}

func (s *ToolService) CreateType(ctx context.Context, in ToolTypeInput) (domain.ToolType, error) {
	t, err := in.toDomain()
	if err != nil {
		return t, err
	}
	if err := s.types.Create(ctx, &t); err != nil {
		return t, err
	}
	return s.types.GetByID(ctx, t.ID)
}

func (s *ToolService) UpdateType(ctx context.Context, id int64, in ToolTypeInput) (domain.ToolType, error) {
	t, err := in.toDomain()
	if err != nil {
		return t, err
	}
	t.ID = id
	if err := s.types.Update(ctx, &t); err != nil {
		return t, err
	}
	return s.types.GetByID(ctx, id)
}

// ---------- tool items ----------

// ToolItemDetail is an item with its maintenance due status.
type ToolItemDetail struct {
	domain.ToolItem
	Maintenance *domain.MaintenanceDue `json:"maintenance"`
}

func (s *ToolService) ListItems(ctx context.Context, f repository.ToolItemFilter) ([]domain.ToolItem, error) {
	if f.Status != nil && !f.Status.Valid() {
		return nil, invalid("unknown status %q", *f.Status)
	}
	return s.items.List(ctx, f)
}

func (s *ToolService) GetItem(ctx context.Context, id int64) (ToolItemDetail, error) {
	it, err := s.items.GetByID(ctx, id)
	if err != nil {
		return ToolItemDetail{}, err
	}
	d := ToolItemDetail{ToolItem: it}
	usage, err := s.maintenance.Usage(ctx, &id)
	if err != nil {
		return d, err
	}
	if len(usage) == 1 { // retired items have no usage row
		due := domain.CheckDue(usage[0], s.now())
		d.Maintenance = &due
	}
	return d, nil
}

type AddItemsInput struct {
	Count      int
	Condition  domain.Condition // default NEW
	AcquiredAt *time.Time       // default now
}

// AddItems adds Count new, individually identified copies of a tool type.
func (s *ToolService) AddItems(ctx context.Context, typeID int64, in AddItemsInput) ([]domain.ToolItem, error) {
	if in.Count == 0 {
		in.Count = 1
	}
	if in.Count < 1 || in.Count > 100 {
		return nil, invalid("count must be between 1 and 100")
	}
	if in.Condition == "" {
		in.Condition = domain.ConditionNew
	}
	if !in.Condition.Valid() || in.Condition == domain.ConditionDamaged {
		return nil, invalid("condition must be NEW, GOOD, FAIR or POOR")
	}
	acquired := s.now()
	if in.AcquiredAt != nil {
		acquired = *in.AcquiredAt
	}
	var out []domain.ToolItem
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.types.GetByID(ctx, typeID); err != nil {
			return err
		}
		for i := 0; i < in.Count; i++ {
			it := domain.ToolItem{ToolTypeID: typeID, Status: domain.ToolAvailable,
				CurrentCondition: in.Condition, AcquiredAt: acquired}
			if err := s.items.Create(ctx, &it); err != nil {
				return err
			}
			out = append(out, it)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

type UpdateItemInput struct {
	Status    *domain.ToolStatus
	Condition *domain.Condition
}

// UpdateItem changes an item's status or condition by hand.
//
// RENTED is set only by handover/return, and MAINTENANCE only by starting a
// maintenance record. A rented item cannot be changed here. An item with an open
// maintenance record can only be retired (or its condition updated).
func (s *ToolService) UpdateItem(ctx context.Context, id int64, in UpdateItemInput) (domain.ToolItem, error) {
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		it, err := s.items.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if it.Status == domain.ToolRented {
			return fmt.Errorf("%w: item is rented; record the return first", domain.ErrInvalidState)
		}
		status, cond := it.Status, it.CurrentCondition
		if in.Condition != nil {
			if !in.Condition.Valid() {
				return invalid("unknown condition %q", *in.Condition)
			}
			cond = *in.Condition
		}
		if in.Status != nil && *in.Status != it.Status {
			switch *in.Status {
			case domain.ToolAvailable, domain.ToolBroken, domain.ToolRetired:
			case domain.ToolRented:
				return invalid("RENTED is set by handing the tool over")
			case domain.ToolMaintenance:
				return invalid("start a maintenance record to put the item into MAINTENANCE")
			default:
				return invalid("unknown status %q", *in.Status)
			}
			if it.Status == domain.ToolRetired {
				return fmt.Errorf("%w: item is retired", domain.ErrInvalidState)
			}
			if it.Status == domain.ToolMaintenance && *in.Status != domain.ToolRetired {
				return fmt.Errorf("%w: item has open maintenance; complete it instead", domain.ErrInvalidState)
			}
			status = *in.Status
		}
		if status == domain.ToolAvailable && cond == domain.ConditionDamaged {
			return invalid("a DAMAGED item cannot be AVAILABLE")
		}
		return s.items.UpdateState(ctx, id, status, cond)
	})
	if err != nil {
		return domain.ToolItem{}, err
	}
	return s.items.GetByID(ctx, id)
}

// ---------- availability ----------

type Availability struct {
	ToolTypeID       int64                 `json:"tool_type_id"`
	From             time.Time             `json:"from"`
	To               time.Time             `json:"to"`
	BookableItemIDs  []int64               `json:"bookable_item_ids"`  // not in maintenance, broken or retired
	AvailableItemIDs []int64               `json:"available_item_ids"` // free for the whole [from, to)
	Busy             []domain.BusyInterval `json:"busy"`
}

// Availability lists which items of a type are taken when in [from, to), so the
// frontend can grey out days where no item is free.
func (s *ToolService) Availability(ctx context.Context, typeID int64, from, to time.Time) (Availability, error) {
	if !to.After(from) {
		return Availability{}, invalid("to must be after from")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return Availability{}, invalid("the period can be at most one year")
	}
	if err := s.status.Refresh(ctx); err != nil {
		return Availability{}, err
	}
	if _, err := s.types.GetByID(ctx, typeID); err != nil {
		return Availability{}, err
	}
	items, err := s.items.List(ctx, repository.ToolItemFilter{ToolTypeID: &typeID})
	if err != nil {
		return Availability{}, err
	}
	busy, err := s.reservations.BusyIntervals(ctx, typeID, from, to)
	if err != nil {
		return Availability{}, err
	}
	taken := map[int64]bool{}
	for _, b := range busy {
		taken[b.ToolItemID] = true
	}
	a := Availability{ToolTypeID: typeID, From: from, To: to, Busy: busy,
		BookableItemIDs: []int64{}, AvailableItemIDs: []int64{}}
	for _, it := range items {
		if !it.Status.Bookable() {
			continue
		}
		a.BookableItemIDs = append(a.BookableItemIDs, it.ID)
		if !taken[it.ID] {
			a.AvailableItemIDs = append(a.AvailableItemIDs, it.ID)
		}
	}
	return a, nil
}
