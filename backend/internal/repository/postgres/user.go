package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type UserRepo struct{ db *DB }

func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

var _ repository.UserRepository = (*UserRepo)(nil)

const userSelect = `SELECT id, username, email, password_hash, role, customer_id, created_at FROM users`

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CustomerID, &u.CreatedAt)
	return u, err
}

func (r *UserRepo) List(ctx context.Context, role *domain.Role) ([]domain.User, error) {
	q := userSelect
	var args []any
	if role != nil {
		q += " WHERE role = $1"
		args = append(args, string(*role))
	}
	q += " ORDER BY id"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list users")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.User, error) { return scanUser(row) })
	return out, mapErr(err, "list users")
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	u, err := scanUser(r.db.q(ctx).QueryRow(ctx, userSelect+" WHERE id = $1", id))
	return u, mapErr(err, "get user")
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	u, err := scanUser(r.db.q(ctx).QueryRow(ctx, userSelect+" WHERE username = $1", username))
	return u, mapErr(err, "get user")
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, customer_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		u.Username, u.Email, u.PasswordHash, u.Role, u.CustomerID).Scan(&u.ID, &u.CreatedAt)
	return mapErr(err, "create user")
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	tag, err := r.db.q(ctx).Exec(ctx,
		`UPDATE users SET email = $2, role = $3, customer_id = $4 WHERE id = $1`,
		u.ID, u.Email, u.Role, u.CustomerID)
	if err != nil {
		return mapErr(err, "update user")
	}
	return expectOne(tag, "update user")
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, hash string) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, id, hash)
	if err != nil {
		return mapErr(err, "update password")
	}
	return expectOne(tag, "update password")
}
