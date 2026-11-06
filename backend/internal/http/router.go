package http

import (
	_ "rentzy/backend/docs"
	"rentzy/backend/internal/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewRouter wires all HTTP routes and returns the configured Gin engine.
func NewRouter(pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	health := handler.NewHealthHandler(pool)

	r.GET("/healthz", health.Healthz)

	// Business endpoints will be registered here.
	// api := r.Group("/api") 
	// api.GET("/tools", toolHandler.List) 
	// api.POST("/tools", toolHandler.Create) 
	
	return r
}
