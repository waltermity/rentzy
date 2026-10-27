package main

import (
	"context"
	"log"
	"os"
	"rentzy/backend/internal/db"
	apphttp "rentzy/backend/internal/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title           Rentzy API
// @version         1.0
// @description     Tool rental administration system API.
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("loading .env file", err)
	}

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)

	pool, err := db.NewPool(context.Background(), dsn)

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	router := apphttp.NewRouter(pool)

	log.Printf("listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("starting server: ", err)
	}

	log.Println("db connected ok")
}
