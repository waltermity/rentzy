package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"rentzy/backend/internal/auth"
	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type AuthService struct {
	tx        repository.TxManager
	users     repository.UserRepository
	customers repository.CustomerRepository
	secret    []byte
	ttl       time.Duration
	now       Clock
}

func NewAuthService(tx repository.TxManager, users repository.UserRepository, customers repository.CustomerRepository,
	secret []byte, ttl time.Duration, now Clock) *AuthService {
	return &AuthService{tx: tx, users: users, customers: customers, secret: secret, ttl: ttl, now: now}
}

type claims struct {
	Role       domain.Role `json:"role"`
	CustomerID *int64      `json:"customer_id,omitempty"`
	jwt.RegisteredClaims
}

// Session is what login and register return.
type Session struct {
	Token     string           `json:"token"`
	ExpiresAt time.Time        `json:"expires_at"`
	User      domain.User      `json:"user"`
	Customer  *domain.Customer `json:"customer,omitempty"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (Session, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		return Session{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if err := auth.CheckPassword(u.PasswordHash, password); err != nil {
		return Session{}, domain.ErrInvalidCredentials
	}
	return s.session(ctx, u)
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
	Name     string
	Phone    string
	Address  string
}

// Register creates a CUSTOMER user and its customer record in one transaction.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (Session, error) {
	var u domain.User
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		c, err := newCustomer(CustomerInput{Name: in.Name, Email: in.Email, Phone: in.Phone, Address: in.Address})
		if err != nil {
			return err
		}
		nu, err := newUser(in.Username, in.Email, in.Password, domain.RoleCustomer)
		if err != nil {
			return err
		}
		if err := s.customers.Create(ctx, &c); err != nil {
			return err
		}
		nu.CustomerID = &c.ID
		if err := s.users.Create(ctx, &nu); err != nil {
			return err
		}
		u = nu
		return nil
	})
	if err != nil {
		return Session{}, err
	}
	return s.session(ctx, u)
}

// Me returns the user and, for customers, their customer record.
func (s *AuthService) Me(ctx context.Context, p domain.Principal) (domain.User, *domain.Customer, error) {
	u, err := s.users.GetByID(ctx, p.UserID)
	if err != nil {
		return domain.User{}, nil, err
	}
	if u.CustomerID == nil {
		return u, nil, nil
	}
	c, err := s.customers.GetByID(ctx, *u.CustomerID)
	if err != nil {
		return domain.User{}, nil, err
	}
	return u, &c, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, p domain.Principal, oldPassword, newPassword string) error {
	u, err := s.users.GetByID(ctx, p.UserID)
	if err != nil {
		return err
	}
	if err := auth.CheckPassword(u.PasswordHash, oldPassword); err != nil {
		return domain.ErrInvalidCredentials
	}
	hash, err := hashValidPassword(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, u.ID, hash)
}

func (s *AuthService) session(ctx context.Context, u domain.User) (Session, error) {
	now := s.now()
	exp := now.Add(s.ttl)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role:       u.Role,
		CustomerID: u.CustomerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	})
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return Session{}, fmt.Errorf("auth: sign token: %w", err)
	}
	sess := Session{Token: signed, ExpiresAt: exp, User: u}
	if u.CustomerID != nil {
		c, err := s.customers.GetByID(ctx, *u.CustomerID)
		if err != nil {
			return Session{}, err
		}
		sess.Customer = &c
	}
	return sess, nil
}

// ParseToken validates a JWT and returns the caller.
func (s *AuthService) ParseToken(token string) (domain.Principal, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.now))
	if err != nil {
		return domain.Principal{}, fmt.Errorf("%w: %v", domain.ErrUnauthorized, err)
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || !c.Role.Valid() {
		return domain.Principal{}, fmt.Errorf("%w: malformed token", domain.ErrUnauthorized)
	}
	return domain.Principal{UserID: id, Role: c.Role, CustomerID: c.CustomerID}, nil
}

const minPasswordLen = 8

func hashValidPassword(pw string) (string, error) {
	if len(pw) < minPasswordLen {
		return "", invalid("password must be at least %d characters", minPasswordLen)
	}
	return auth.HashPassword(pw)
}

func newUser(username, email, password string, role domain.Role) (domain.User, error) {
	username, err := requireText("username", username)
	if err != nil {
		return domain.User{}, err
	}
	if len(username) < 3 {
		return domain.User{}, invalid("username must be at least 3 characters")
	}
	if !validEmail(email) {
		return domain.User{}, invalid("email is not valid")
	}
	if !role.Valid() {
		return domain.User{}, invalid("role must be ADMIN, STAFF or CUSTOMER")
	}
	hash, err := hashValidPassword(password)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{Username: username, Email: email, PasswordHash: hash, Role: role}, nil
}
