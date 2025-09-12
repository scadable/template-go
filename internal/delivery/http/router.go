package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"template-go/internal/delivery/http/routes"
)

func NewRouter(serviceName string) http.Handler {
	r := chi.NewRouter()

	// OTel Middleware
	// This should be the first middleware
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, serviceName)
	})

	// Common middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)    // logs every request
	r.Use(middleware.Recoverer) // recovers from panics

	// Serve metrics at /metrics
	r.Handle("/metrics", promhttp.Handler())

	// Serve the swagger documentation at /docs/index.html
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	// Attach root route
	r.Mount("/", routes.RootRoutes())

	return r
}
