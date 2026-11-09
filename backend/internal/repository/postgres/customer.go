package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type CustomerRepo struct{ db *DB }

func NewCustomerRepo(db *DB) *CustomerRepo { return &CustomerRepo{db: db} }

var _ repository.CustomerRepository = (*CustomerRepo)(nil)

const customerSelect = `
SELECT id, name, email, COALESCE(phone, ''), COALESCE(address, ''), created_at FROM customers`

func scanCustomer(row pgx.Row) (domain.Customer, error) {
	var c domain.Customer
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.CreatedAt)
	return c, err
}

func (r *CustomerRepo) List(ctx context.Context, search string) ([]domain.Customer, error) {
	q := customerSelect
	var args []any
	if s := strings.TrimSpace(search); s != "" {
		q += " WHERE name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1"
		args = append(args, "%"+s+"%")
	}
	q += " ORDER BY name, id"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list customers")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.Customer, error) { return scanCustomer(row) })
	return out, mapErr(err, "list customers")
}

func (r *CustomerRepo) GetByID(ctx context.Context, id int64) (domain.Customer, error) {
	c, err := scanCustomer(r.db.q(ctx).QueryRow(ctx, customerSelect+" WHERE id = $1", id))
	return c, mapErr(err, "get customer")
}

func (r *CustomerRepo) Create(ctx context.Context, c *domain.Customer) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO customers (name, email, phone, address) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		c.Name, c.Email, nullIfEmpty(c.Phone), nullIfEmpty(c.Address)).Scan(&c.ID, &c.CreatedAt)
	return mapErr(err, "create customer")
}

func (r *CustomerRepo) Update(ctx context.Context, c *domain.Customer) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		UPDATE customers SET name = $2, email = $3, phone = $4, address = $5 WHERE id = $1`,
		c.ID, c.Name, c.Email, nullIfEmpty(c.Phone), nullIfEmpty(c.Address))
	if err != nil {
		return mapErr(err, "update customer")
	}
	return expectOne(tag, "update customer")
}
