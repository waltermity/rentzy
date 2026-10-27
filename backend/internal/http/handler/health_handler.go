package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"rentzy/backend/internal/http/respond"
)

// HealthHandler groups the health-related endpoints.
type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	if err := h.pool.Ping(c.Request.Context()); err != nil {
		log.Printf("healthz: db ping failed: %v", err)
		respond.Error(c, http.StatusServiceUnavailable, "db_unreachable", "database is unreachable")
		return
	}
	respond.OK(c, gin.H{"status": "ok"})
}
