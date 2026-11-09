package service

import (
	"context"
	"strings"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type CustomerService struct {
	customers repository.CustomerRepository
}

func NewCustomerService(customers repository.CustomerRepository) *CustomerService {
	return &CustomerService{customers: customers}
}

type CustomerInput struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

func newCustomer(in CustomerInput) (domain.Customer, error) {
	name, err := requireText("name", in.Name)
	if err != nil {
		return domain.Customer{}, err
	}
	email := strings.TrimSpace(in.Email)
	if !validEmail(email) {
		return domain.Customer{}, invalid("email is not valid")
	}
	return domain.Customer{Name: name, Email: email,
		Phone: strings.TrimSpace(in.Phone), Address: strings.TrimSpace(in.Address)}, nil
}

func (s *CustomerService) List(ctx context.Context, search string) ([]domain.Customer, error) {
	return s.customers.List(ctx, search)
}

// Get returns a customer. A CUSTOMER caller may only read their own record.
func (s *CustomerService) Get(ctx context.Context, p domain.Principal, id int64) (domain.Customer, error) {
	if !p.IsStaffOrAdmin() && (p.CustomerID == nil || *p.CustomerID != id) {
		return domain.Customer{}, domain.ErrNotFound
	}
	return s.customers.GetByID(ctx, id)
}

func (s *CustomerService) Create(ctx context.Context, in CustomerInput) (domain.Customer, error) {
	c, err := newCustomer(in)
	if err != nil {
		return c, err
	}
	return c, s.customers.Create(ctx, &c)
}

func (s *CustomerService) Update(ctx context.Context, id int64, in CustomerInput) (domain.Customer, error) {
	c, err := newCustomer(in)
	if err != nil {
		return c, err
	}
	c.ID = id
	if err := s.customers.Update(ctx, &c); err != nil {
		return c, err
	}
	return s.customers.GetByID(ctx, id)
}
