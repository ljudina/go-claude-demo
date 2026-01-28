package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteRegistrar interface {
	RegisterRoutes(r chi.Router)
}

type RouteHandler struct {
	router *chi.Mux
}

func NewRouteHandler() *RouteHandler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return &RouteHandler{router: r}
}

func (h *RouteHandler) Router() chi.Router {
	return h.router
}

func (h *RouteHandler) AddRoute(reg RouteRegistrar) {
	reg.RegisterRoutes(h.router)
}

func (h *RouteHandler) Start(port string) {
	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, h.router); err != nil {
		log.Fatal("server error:", err)
	}
}

