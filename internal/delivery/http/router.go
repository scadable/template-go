package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"

	"template-go/internal/delivery/http/routes"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

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
