package main

import (
	"context"
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

func ctx() context.Context {
	return context.Background()
}

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

	// Ensure nav-items policies exist (for existing databases)
	if err := authz.EnsureNavItemPolicies(enforcer); err != nil {
		log.Fatal("failed to ensure nav-item policies:", err)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	navItemRepo := repository.NewNavItemRepository(db)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)
	navItemService := service.NewNavItemService(navItemRepo, roleService)

	// Seed default navigation items if none exist
	if err := navItemService.SeedDefaultNavItems(ctx()); err != nil {
		log.Fatal("failed to seed default nav items:", err)
	}

	// Auth handler
	sessionSecret := os.Getenv("SESSION_SECRET")
	authHandler := handler.NewAuthHandler(userService, roleRepo, navItemRepo, sessionSecret)

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

	webRoleHandler := webhandler.NewWebRoleHandler(roleService, navItemService, authHandler)
	webPolicyHandler := webhandler.NewWebPolicyHandler(enforcer, roleService, navItemService, authHandler)
	webUserHandler := webhandler.NewWebUserHandler(userService, roleService, navItemService, authHandler)
	webNavItemHandler := webhandler.NewWebNavItemHandler(navItemService, roleService, authHandler)

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

		// Nav Items web
		r.Route("/nav-items", func(r chi.Router) {
			r.With(authzMiddleware.RequirePermission("read")).Get("/", webNavItemHandler.ListNavItems)
			r.With(authzMiddleware.RequirePermission("write")).Get("/new", webNavItemHandler.NewNavItemForm)
			r.With(authzMiddleware.RequirePermission("write")).Post("/", webNavItemHandler.CreateNavItem)
			r.With(authzMiddleware.RequirePermission("read")).Get("/{id}/edit", webNavItemHandler.EditNavItemForm)
			r.With(authzMiddleware.RequirePermission("write")).Put("/{id}", webNavItemHandler.UpdateNavItem)
			r.With(authzMiddleware.RequirePermission("write")).Delete("/{id}", webNavItemHandler.DeleteNavItem)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	routeHandler.Start(port)
}
