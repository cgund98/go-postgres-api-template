package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	otelchimetric "github.com/riandyrn/otelchi/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
)

// Router wraps Chi router with Huma API
type Router struct {
	chiRouter *chi.Mux
	humaAPI   huma.API
}

// NewRouter creates a new router with Chi and Huma
func NewRouter(serverName string, rootLogger *slog.Logger, tracer trace.Tracer, meterProvider *metric.MeterProvider) *Router {
	chiRouter := chi.NewRouter()

	otelChiConfig := otelchimetric.NewBaseConfig(serverName, otelchimetric.WithMeterProvider(meterProvider))
	chiRouter.Use(
		otelchi.Middleware(serverName, otelchi.WithChiRoutes(chiRouter)),
		otelchimetric.NewServerRequestDuration(otelChiConfig),
		otelchimetric.NewServerActiveRequests(otelChiConfig),
		otelchimetric.NewServerRequestBodySize(otelChiConfig),
		otelchimetric.NewServerResponseBodySize(otelChiConfig),
	)

	// Add request logging middleware
	chiRouter.Use(RequestLogger(rootLogger))
	chiRouter.Use(TracerSetter(tracer))

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
