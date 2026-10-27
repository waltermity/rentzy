package http

import (
	"rentzy/backend/internal/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter wires all HTTP routes and returns the configured Gin engine.
func NewRouter(pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	health := handler.NewHealthHandler(pool)

	r.GET("/healthz", health.Healthz)

	return r
}
