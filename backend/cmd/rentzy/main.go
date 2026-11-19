package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"

	"rentzy/backend/internal/db"
	apphttp "rentzy/backend/internal/http"
	"rentzy/backend/internal/repository/postgres"
	"rentzy/backend/internal/service"
)

// @title                       Rentzy API
// @version                     1.0
// @description                 Tool rental administration system API.
// @BasePath                    /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 "Bearer <token>" from POST /api/auth/login
func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("loading .env file", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be set to at least 32 characters")
	}
	port := envOr("PORT", "8080")
	jwtTTL := envDuration("JWT_TTL", 24*time.Hour)
	grace := time.Duration(envInt("RESERVATION_GRACE_HOURS", 24)) * time.Hour
	refreshEvery := envDuration("STATUS_REFRESH_INTERVAL", time.Minute)

	// Money is decimal.Decimal; send it as a JSON number, not a string.
	decimal.MarshalJSONWithoutQuotes = true

	gin.SetMode(gin.ReleaseMode)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Println("db connected ok")

	// Repositories (PostgreSQL implementations of the repository interfaces).
	store := postgres.NewDB(pool)
	categories := postgres.NewCategoryRepo(store)
	toolTypes := postgres.NewToolTypeRepo(store)
	toolItems := postgres.NewToolItemRepo(store)
	customers := postgres.NewCustomerRepo(store)
	users := postgres.NewUserRepo(store)
	reservations := postgres.NewReservationRepo(store)
	rentals := postgres.NewRentalRepo(store)
	maintenance := postgres.NewMaintenanceRepo(store)
	reports := postgres.NewReportRepo(store)

	// Services.
	now := service.Clock(time.Now)
	status := service.NewStatusService(rentals, reservations, grace, now)
	services := apphttp.Services{
		Auth:        service.NewAuthService(store, users, customers, []byte(jwtSecret), jwtTTL, now),
		Users:       service.NewUserService(store, users, customers),
		Customers:   service.NewCustomerService(customers),
		Tools:       service.NewToolService(store, categories, toolTypes, toolItems, reservations, maintenance, status, now),
		Reservation: service.NewReservationService(store, reservations, toolItems, customers, status, now),
		Rentals:     service.NewRentalService(store, rentals, reservations, toolItems, toolTypes, status, now),
		Maintenance: service.NewMaintenanceService(store, maintenance, toolItems, now),
		Reports:     service.NewReportService(reports, now),
	}

	// OVERDUE rentals and EXPIRED reservations, in the background.
	go status.Run(ctx, refreshEvery)

	srv := &http.Server{Addr: ":" + port, Handler: apphttp.NewRouter(pool, services)}
	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("starting server: ", err)
		}
	}()

	// Ctrl+C / SIGTERM: stop accepting requests, let running ones finish (max 10s).
	<-ctx.Done()
	stop() // a second Ctrl+C now kills the process immediately
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("server stopped")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		log.Fatalf("%s must be a non-negative integer", key)
	}
	return n
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		log.Fatalf("%s must be a positive duration like 24h or 1m", key)
	}
	return d
}
