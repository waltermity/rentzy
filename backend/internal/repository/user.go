package repository

import (
	"context"

	"rentzy/backend/internal/domain"
)

type UserRepository interface {
	List(ctx context.Context, role *domain.Role) ([]domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	// Update stores email, role and customer_id.
	Update(ctx context.Context, u *domain.User) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
}
