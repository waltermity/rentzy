package repository

import (
	"context"

	"rentzy/backend/internal/domain"
)

type CustomerRepository interface {
	List(ctx context.Context, search string) ([]domain.Customer, error)
	GetByID(ctx context.Context, id int64) (domain.Customer, error)
	Create(ctx context.Context, c *domain.Customer) error
	Update(ctx context.Context, c *domain.Customer) error
}
