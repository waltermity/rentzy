package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/respond"
)

func badRequest(c *gin.Context, format string, args ...any) {
	respond.Err(c, fmt.Errorf("%w: %s", domain.ErrInvalidInput, fmt.Sprintf(format, args...)))
}

// pathID reads a positive int64 path parameter; on failure it writes 400.
func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		badRequest(c, "%s must be a positive integer", name)
		return 0, false
	}
	return id, true
}

func queryInt64(c *gin.Context, name string) (*int64, bool) {
	v := c.Query(name)
	if v == "" {
		return nil, true
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		badRequest(c, "%s must be an integer", name)
		return nil, false
	}
	return &n, true
}

func queryFloat(c *gin.Context, name string, def float64) (float64, bool) {
	v := c.Query(name)
	if v == "" {
		return def, true
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		badRequest(c, "%s must be a number", name)
		return 0, false
	}
	return f, true
}

// queryTime accepts RFC 3339 (2026-11-09T08:00:00+01:00) or a date (2026-11-09, UTC midnight).
func queryTime(c *gin.Context, name string) (*time.Time, bool) {
	v := c.Query(name)
	if v == "" {
		return nil, true
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return &t, true
	}
	if t, err := time.Parse(time.DateOnly, v); err == nil {
		return &t, true
	}
	badRequest(c, "%s must be an RFC 3339 timestamp or YYYY-MM-DD date", name)
	return nil, false
}

// queryList splits a comma-separated parameter: ?status=ACTIVE,OVERDUE
func queryList(c *gin.Context, name string) []string {
	var out []string
	for _, v := range strings.Split(c.Query(name), ",") {
		if v = strings.TrimSpace(strings.ToUpper(v)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		badRequest(c, "invalid request body: %v", err)
		return false
	}
	return true
}
