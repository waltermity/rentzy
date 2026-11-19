package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/respond"
)

// RequireRoles lets the request through only for the given roles. Use after Auth.
func RequireRoles(roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(roles, PrincipalFrom(c).Role) {
			respond.Error(c, http.StatusForbidden, "forbidden", "your role cannot access this resource")
			return
		}
		c.Next()
	}
}
