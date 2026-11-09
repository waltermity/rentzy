package domain

import "time"

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleStaff    Role = "STAFF"
	RoleCustomer Role = "CUSTOMER"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleStaff || r == RoleCustomer
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CustomerID   *int64    `json:"customer_id"` // set only for CUSTOMER users
	CreatedAt    time.Time `json:"created_at"`
}

// Principal is the authenticated caller, taken from the JWT.
type Principal struct {
	UserID     int64
	Role       Role
	CustomerID *int64
}

func (p Principal) IsStaffOrAdmin() bool {
	return p.Role == RoleAdmin || p.Role == RoleStaff
}
