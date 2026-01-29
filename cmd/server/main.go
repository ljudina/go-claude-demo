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
	"claude-test/internal/handler/apihandler"
	"claude-test/internal/handler/webhandler"
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

	apiUserHandler := apihandler.NewApiUserHandler(userService)
	apiRoleHandler := apihandler.NewApiRoleHandler(roleService)

	routeHandler := handler.NewRouteHandler()

	// Public routes (auth)
	routeHandler.AddRoute(authHandler)

	// Create authorization middleware
	roleGetter := authz.NewRoleService(roleService)

	authzMiddleware := authz.NewMiddleware(enforcer, authHandler, roleGetter)

	// API routes with authorization
	routeHandler.Router().Route("/api", func(r chi.Router) {
		r.Use(authzMiddleware.RequireAuth)

		// Users API
		r.Route("/users", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", apiUserHandler.List)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", apiUserHandler.Create)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}", apiUserHandler.GetByID)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", apiUserHandler.Update)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", apiUserHandler.Delete)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{userId}/roles", apiRoleHandler.GetUserRoles)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{userId}/roles", apiRoleHandler.SetUserRoles)
		})

		// Roles API
		r.Route("/roles", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", apiRoleHandler.List)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", apiRoleHandler.Create)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}", apiRoleHandler.GetByID)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", apiRoleHandler.Update)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", apiRoleHandler.Delete)
		})
	})

	webRoleHandler := webhandler.NewWebRoleHandler(roleService, authHandler)
	webPolicyHandler := webhandler.NewWebPolicyHandler(enforcer, roleService, authHandler)
	webUserHandler := webhandler.NewWebUserHandler(userService, roleService, authHandler)

	// Web routes with authorization
	routeHandler.Router().Route("/web", func(r chi.Router) {
		r.Use(authzMiddleware.RequireAuth)

		// Users web
		r.Route("/users", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", webUserHandler.ListUsers)
			r.With(authzMiddleware.RequirePermission("write")).Get("/new", webUserHandler.NewUserForm)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", webUserHandler.CreateUser)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}/edit", webUserHandler.EditUserForm)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", webUserHandler.UpdateUser)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", webUserHandler.DeleteUser)
		})

		// Roles web
		r.Route("/roles", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", webRoleHandler.ListRoles)
			r.With(authzMiddleware.RequirePermission("write")).Get("/new", webRoleHandler.NewRoleForm)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", webRoleHandler.CreateRole)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}/edit", webRoleHandler.EditRoleForm)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", webRoleHandler.UpdateRole)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", webRoleHandler.DeleteRole)
		})

		// Policies web (Admin only)
		r.Route("/policies", func(r chi.Router) {
			r.Use(authzMiddleware.RequireRole("Admin"))
			r.Get("/", webPolicyHandler.ListPolicies)
			r.Get("/new", webPolicyHandler.NewPolicyForm)
			r.Post("/", webPolicyHandler.CreatePolicy)
			r.Delete("/", webPolicyHandler.DeletePolicy)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	routeHandler.Start(port)
}
