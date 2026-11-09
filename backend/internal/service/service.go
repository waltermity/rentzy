// Package service holds the business rules. It depends only on the repository
// interfaces, never on a concrete database.
package service

import (
	"fmt"
	"strings"
	"time"

	"rentzy/backend/internal/domain"
)

// Clock returns the current time; tests can replace it.
type Clock func() time.Time

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalidInput, fmt.Sprintf(format, args...))
}

func requireText(field, v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", invalid("%s is required", field)
	}
	return v, nil
}

func validEmail(v string) bool {
	at := strings.Index(v, "@")
	return at > 0 && at < len(v)-1 && !strings.ContainsAny(v, " \t\n")
}
