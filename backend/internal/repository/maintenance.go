package repository

import (
	"context"

	"rentzy/backend/internal/domain"
)

type MaintenanceFilter struct {
	ToolItemID *int64
	Open       *bool // true: not completed yet, false: completed
}

type MaintenanceRepository interface {
	List(ctx context.Context, f MaintenanceFilter) ([]domain.Maintenance, error)
	GetByID(ctx context.Context, id int64) (domain.Maintenance, error)
	GetByIDForUpdate(ctx context.Context, id int64) (domain.Maintenance, error)
	Create(ctx context.Context, m *domain.Maintenance) error
	// Complete stores completed_at, cost, description and performed_by.
	Complete(ctx context.Context, m *domain.Maintenance) error
	CountOpen(ctx context.Context, toolItemID int64) (int, error)
	// Usage returns the usage since last maintenance for every item that is not
	// RETIRED; toolItemID != nil limits it to one item.
	Usage(ctx context.Context, toolItemID *int64) ([]domain.MaintenanceUsage, error)
}
