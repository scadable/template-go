package http

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"

	"template-go/internal/delivery/http/routes"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Serve the swagger documentation at /docs/index.html
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	// Attach root route
	r.Mount("/", routes.RootRoutes())

	return r
}
