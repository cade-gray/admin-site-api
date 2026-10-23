package main

import (
	"admin-site-api/lib"
	"admin-site-api/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	router := gin.Default()

	// Cross-origin requests are answered here, from an allowlist, because the
	// browser apps sign in with a cookie and credentialed CORS cannot use a
	// wildcard. Caddy must not add Access-Control-* headers of its own.
	corsOrigins := lib.CORSOrigins()
	if corsOrigins == "" {
		log.Println("warning: CORS_ORIGINS is empty, so no browser app can call this API")
	}
	router.Use(lib.CORS(corsOrigins))

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	authDbname := os.Getenv("AUTH_DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	authDsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, authDbname, port, sslmode,
	)
	authDb, err := gorm.Open(postgres.Open(authDsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect auth database")
	}

	routes.RegisterJokeRoutes(router, authDb)
	routes.RegisterPlateRoutes(router, db)
	routes.RegisterAuthRoutes(router, authDb)
	routes.RegisterProjectRoutes(router, db)

	router.Run(":4269")
}
