package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/respond"
)

const principalKey = "principal"

// TokenParser validates a bearer token and returns the caller.
type TokenParser interface {
	ParseToken(token string) (domain.Principal, error)
}

// Auth requires a valid "Authorization: Bearer <jwt>" header and stores the
// caller in the context.
func Auth(p TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			respond.Error(c, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		principal, err := p.ParseToken(strings.TrimSpace(token))
		if err != nil {
			respond.Error(c, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}
		c.Set(principalKey, principal)
		c.Next()
	}
}

// PrincipalFrom returns the caller stored by Auth.
func PrincipalFrom(c *gin.Context) domain.Principal {
	p, _ := c.Get(principalKey)
	principal, _ := p.(domain.Principal)
	return principal
}
