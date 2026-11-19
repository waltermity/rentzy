package service

import (
	"context"
	"strings"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

// UserService is the admin's account management.
type UserService struct {
	tx        repository.TxManager
	users     repository.UserRepository
	customers repository.CustomerRepository
}

func NewUserService(tx repository.TxManager, users repository.UserRepository, customers repository.CustomerRepository) *UserService {
	return &UserService{tx: tx, users: users, customers: customers}
}

type CreateUserInput struct {
	Username string
	Email    string
	Password string
	Role     domain.Role
	// For CUSTOMER users: link an existing customer (CustomerID), or create one
	// (Customer). Exactly one of them is required.
	CustomerID *int64
	Customer   *CustomerInput
}

func (s *UserService) List(ctx context.Context, role *domain.Role) ([]domain.User, error) {
	return s.users.List(ctx, role)
}

func (s *UserService) Get(ctx context.Context, id int64) (domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (domain.User, error) {
	u, err := newUser(in.Username, in.Email, in.Password, in.Role)
	if err != nil {
		return u, err
	}
	if u.Role != domain.RoleCustomer {
		if in.CustomerID != nil || in.Customer != nil {
			return u, invalid("only CUSTOMER users can be linked to a customer")
		}
		return u, s.users.Create(ctx, &u)
	}
	if (in.CustomerID == nil) == (in.Customer == nil) {
		return u, invalid("CUSTOMER users need either customer_id or customer")
	}
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if in.CustomerID != nil {
			if _, err := s.customers.GetByID(ctx, *in.CustomerID); err != nil {
				return err
			}
			u.CustomerID = in.CustomerID
		} else {
			c, err := newCustomer(*in.Customer)
			if err != nil {
				return err
			}
			if err := s.customers.Create(ctx, &c); err != nil {
				return err
			}
			u.CustomerID = &c.ID
		}
		return s.users.Create(ctx, &u)
	})
	return u, err
}

type UpdateUserInput struct {
	Email    *string
	Role     *domain.Role
	Password *string
}

// Update changes email, password or role. Roles can change between ADMIN and
// STAFF only: a CUSTOMER account is tied to its customer record.
func (s *UserService) Update(ctx context.Context, id int64, in UpdateUserInput) (domain.User, error) {
	var out domain.User
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.users.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if in.Email != nil {
			e := strings.TrimSpace(*in.Email)
			if !validEmail(e) {
				return invalid("email is not valid")
			}
			u.Email = e
		}
		if in.Role != nil && *in.Role != u.Role {
			if !in.Role.Valid() {
				return invalid("role must be ADMIN, STAFF or CUSTOMER")
			}
			if u.Role == domain.RoleCustomer || *in.Role == domain.RoleCustomer {
				return invalid("role can only change between ADMIN and STAFF")
			}
			u.Role = *in.Role
		}
		if err := s.users.Update(ctx, &u); err != nil {
			return err
		}
		if in.Password != nil {
			hash, err := hashValidPassword(*in.Password)
			if err != nil {
				return err
			}
			if err := s.users.UpdatePassword(ctx, u.ID, hash); err != nil {
				return err
			}
		}
		out = u
		return nil
	})
	return out, err
}
