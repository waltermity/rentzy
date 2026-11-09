package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "rentzy/backend/docs"
	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/handler"
	"rentzy/backend/internal/http/middleware"
	"rentzy/backend/internal/service"
)

// Services are the business services the routes call.
type Services struct {
	Auth        *service.AuthService
	Users       *service.UserService
	Customers   *service.CustomerService
	Tools       *service.ToolService
	Reservation *service.ReservationService
	Rentals     *service.RentalService
	Maintenance *service.MaintenanceService
	Reports     *service.ReportService
}

// NewRouter wires all HTTP routes and returns the configured Gin engine.
//
// Access (RBAC):
//   - public: health, docs, login, register
//   - any logged-in user: own profile, catalog, availability, own reservations
//     and rentals (customers see only their own)
//   - STAFF + ADMIN: customers (read), tool items, handover/return,
//     maintenance (read), popular-tools report
//   - ADMIN: catalog and item management, customer and user management,
//     maintenance (write), utilization/revenue and procurement reports
func NewRouter(pool *pgxpool.Pool, s Services) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logging())

	// Swagger UI: /docs and /docs/ redirect to /docs/index.html.
	swaggerUI := ginSwagger.WrapHandler(swaggerFiles.Handler)
	toIndex := func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/docs/index.html") }
	r.GET("/docs", toIndex)
	r.GET("/docs/*any", func(c *gin.Context) {
		if c.Param("any") == "/" {
			toIndex(c)
			return
		}
		swaggerUI(c)
	})
	r.GET("/healthz", handler.NewHealthHandler(pool).Healthz)

	authH := handler.NewAuthHandler(s.Auth)
	userH := handler.NewUserHandler(s.Users)
	customerH := handler.NewCustomerHandler(s.Customers)
	toolH := handler.NewToolHandler(s.Tools)
	reservationH := handler.NewReservationHandler(s.Reservation)
	rentalH := handler.NewRentalHandler(s.Rentals)
	maintH := handler.NewMaintenanceHandler(s.Maintenance)
	reportH := handler.NewReportHandler(s.Reports)

	api := r.Group("/api")
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/register", authH.Register)

	authed := api.Group("", middleware.Auth(s.Auth))
	staff := authed.Group("", middleware.RequireRoles(domain.RoleStaff, domain.RoleAdmin))
	admin := authed.Group("", middleware.RequireRoles(domain.RoleAdmin))

	// any logged-in user
	authed.GET("/auth/me", authH.Me)
	authed.POST("/auth/password", authH.ChangePassword)
	authed.GET("/categories", toolH.ListCategories)
	authed.GET("/tool-types", toolH.ListTypes)
	authed.GET("/tool-types/:id", toolH.GetType)
	authed.GET("/tool-types/:id/availability", toolH.Availability)
	authed.GET("/customers/:id", customerH.Get)
	authed.POST("/reservations", reservationH.Create)
	authed.GET("/reservations", reservationH.List)
	authed.GET("/reservations/:id", reservationH.Get)
	authed.POST("/reservations/:id/cancel", reservationH.Cancel)
	authed.GET("/rentals", rentalH.List)
	authed.GET("/rentals/:id", rentalH.Get)

	// staff and admin
	staff.GET("/customers", customerH.List)
	staff.GET("/tool-items", toolH.ListItems)
	staff.GET("/tool-items/:id", toolH.GetItem)
	staff.GET("/tool-items/:id/maintenance", maintH.History)
	staff.POST("/rentals", rentalH.Handover)
	staff.POST("/rentals/:id/return", rentalH.Return)
	staff.POST("/rentals/:id/cancel", rentalH.Cancel)
	staff.GET("/maintenance", maintH.List)
	staff.GET("/maintenance/due", maintH.Due)
	staff.GET("/maintenance/:id", maintH.Get)
	staff.GET("/reports/popular", reportH.Popular)

	// admin only
	admin.POST("/categories", toolH.CreateCategory)
	admin.PUT("/categories/:id", toolH.UpdateCategory)
	admin.POST("/tool-types", toolH.CreateType)
	admin.PUT("/tool-types/:id", toolH.UpdateType)
	admin.POST("/tool-types/:id/items", toolH.AddItems)
	admin.PATCH("/tool-items/:id", toolH.UpdateItem)
	admin.POST("/tool-items/:id/maintenance", maintH.Start)
	admin.POST("/maintenance/:id/complete", maintH.Complete)
	admin.POST("/customers", customerH.Create)
	admin.PUT("/customers/:id", customerH.Update)
	admin.GET("/users", userH.List)
	admin.POST("/users", userH.Create)
	admin.GET("/users/:id", userH.Get)
	admin.PATCH("/users/:id", userH.Update)
	admin.GET("/reports/utilization", reportH.Utilization)
	admin.GET("/reports/procurement", reportH.Procurement)

	return r
}
