package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"claude-test/internal/database"
	"claude-test/internal/domain"
	"claude-test/internal/handler"
	"claude-test/internal/repository"
	"claude-test/internal/service"
)

func init() {
	// Register types for session serialization
	gob.Register(&domain.User{})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("failed to run migrations:", err)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)

	// Auth handler
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "default-secret-change-me"
	}
	authHandler := handler.NewAuthHandler(userService, roleRepo, sessionSecret)

	// Initialize Google OAuth
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}
	authHandler.InitGoth(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		appURL,
	)

	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	webHandler := handler.NewWebHandler(userService, roleService)
	roleWebHandler := handler.NewRoleWebHandler(roleService)

	routeHandler := handler.NewRouteHandler()
	routeHandler.AddRoute(authHandler)
	routeHandler.AddRoute(userHandler)
	routeHandler.AddRoute(roleHandler)
	routeHandler.AddRoute(webHandler)
	routeHandler.AddRoute(roleWebHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	routeHandler.Start(port)
}
