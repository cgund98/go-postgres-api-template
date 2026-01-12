package presentation

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// Router wraps Chi router with Huma API
type Router struct {
	chiRouter *chi.Mux
	humaAPI   huma.API
}

// NewRouter creates a new router with Chi and Huma
func NewRouter() *Router {
	chiRouter := chi.NewRouter()

	// Add request logging middleware
	chiRouter.Use(RequestLogger())

	// Create Huma API adapter for Chi
	// DefaultConfig sets up /openapi.json, /docs, and /schemas endpoints
	config := huma.DefaultConfig("My API", "1.0.0")
	humaAPI := humachi.New(chiRouter, config)

	return &Router{
		chiRouter: chiRouter,
		humaAPI:   humaAPI,
	}
}

// ChiRouter returns the underlying Chi router (for mounting non-Huma routes)
func (r *Router) ChiRouter() *chi.Mux {
	return r.chiRouter
}

// HumaAPI returns the Huma API instance
func (r *Router) HumaAPI() huma.API {
	return r.humaAPI
}

// ServeHTTP implements http.Handler for the router
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chiRouter.ServeHTTP(w, req)
}
