package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"claude-test/internal/authz"
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

	// Initialize Casbin enforcer
	enforcer, err := authz.NewEnforcer(db)
	if err != nil {
		log.Fatal("failed to create enforcer:", err)
	}

	// Setup default policies if none exist
	if err := authz.SetupDefaultPolicies(enforcer); err != nil {
		log.Fatal("failed to setup default policies:", err)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)

	// Auth handler
	sessionSecret := os.Getenv("SESSION_SECRET")
	authHandler := handler.NewAuthHandler(userService, roleRepo, sessionSecret)

	// Initialize Google OAuth
	appURL := os.Getenv("APP_URL")
	authHandler.InitGoth(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		appURL,
	)

	// Create authorization middleware
	roleGetter := authz.NewRoleService(roleService)
	authzMiddleware := authz.NewMiddleware(enforcer, authHandler, roleGetter)

	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	webHandler := handler.NewWebHandler(userService, roleService, authHandler)
	roleWebHandler := handler.NewRoleWebHandler(roleService, authHandler)
	policyWebHandler := handler.NewPolicyWebHandler(enforcer, roleService, authHandler)

	routeHandler := handler.NewRouteHandler()

	// Public routes (auth)
	routeHandler.AddRoute(authHandler)

	// API routes with authorization
	routeHandler.Router().Route("/api", func(r chi.Router) {
		r.Use(authzMiddleware.RequireAuth)

		// Users API
		r.Route("/users", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", userHandler.List)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", userHandler.Create)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}", userHandler.GetByID)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", userHandler.Update)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", userHandler.Delete)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{userId}/roles", roleHandler.GetUserRoles)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{userId}/roles", roleHandler.SetUserRoles)
		})

		// Roles API
		r.Route("/roles", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", roleHandler.List)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", roleHandler.Create)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}", roleHandler.GetByID)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", roleHandler.Update)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", roleHandler.Delete)
		})
	})

	// Web routes with authorization
	routeHandler.Router().Route("/web", func(r chi.Router) {
		r.Use(authzMiddleware.RequireAuth)

		// Users web
		r.Route("/users", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", webHandler.ListUsers)
			r.With(authzMiddleware.RequirePermission("write")).Get("/new", webHandler.NewUserForm)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", webHandler.CreateUser)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}/edit", webHandler.EditUserForm)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", webHandler.UpdateUser)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", webHandler.DeleteUser)
		})

		// Roles web
		r.Route("/roles", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", roleWebHandler.ListRoles)
			r.With(authzMiddleware.RequirePermission("write")).Get("/new", roleWebHandler.NewRoleForm)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", roleWebHandler.CreateRole)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}/edit", roleWebHandler.EditRoleForm)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", roleWebHandler.UpdateRole)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", roleWebHandler.DeleteRole)
		})

		// Policies web (Admin only)
		r.Route("/policies", func(r chi.Router) {
			r.Use(authzMiddleware.RequireRole("Admin"))
			r.Get("/", policyWebHandler.ListPolicies)
			r.Get("/new", policyWebHandler.NewPolicyForm)
			r.Post("/", policyWebHandler.CreatePolicy)
			r.Delete("/", policyWebHandler.DeletePolicy)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	routeHandler.Start(port)
}
