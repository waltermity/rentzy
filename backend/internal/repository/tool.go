package repository

import (
	"context"

	"rentzy/backend/internal/domain"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id int64) (domain.Category, error)
	Create(ctx context.Context, c *domain.Category) error
	Update(ctx context.Context, c *domain.Category) error
}

type ToolTypeFilter struct {
	CategoryID *int64
	Search     string // matches name, manufacturer or model (case-insensitive)
}

type ToolTypeRepository interface {
	List(ctx context.Context, f ToolTypeFilter) ([]domain.ToolType, error)
	GetByID(ctx context.Context, id int64) (domain.ToolType, error)
	Create(ctx context.Context, t *domain.ToolType) error
	Update(ctx context.Context, t *domain.ToolType) error
}

type ToolItemFilter struct {
	ToolTypeID *int64
	Status     *domain.ToolStatus
}

type ToolItemRepository interface {
	List(ctx context.Context, f ToolItemFilter) ([]domain.ToolItem, error)
	GetByID(ctx context.Context, id int64) (domain.ToolItem, error)
	// GetByIDForUpdate locks the item row until the transaction ends.
	GetByIDForUpdate(ctx context.Context, id int64) (domain.ToolItem, error)
	// ListByTypeForUpdate locks all items of a type (ordered by id).
	ListByTypeForUpdate(ctx context.Context, toolTypeID int64) ([]domain.ToolItem, error)
	Create(ctx context.Context, item *domain.ToolItem) error
	UpdateState(ctx context.Context, id int64, status domain.ToolStatus, condition domain.Condition) error
}
