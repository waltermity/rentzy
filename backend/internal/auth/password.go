// Package auth hashes and checks user passwords.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns a bcrypt hash of password at bcrypt.DefaultCost, with a
// new random salt. Store only the hash, never the password itself.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword returns nil if password matches hash. A wrong password gives
// bcrypt.ErrMismatchedHashAndPassword; a hash that is not valid bcrypt gives
// another error. It also accepts the hashes that pgcrypto's crypt() makes for
// the dev seed data.
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
