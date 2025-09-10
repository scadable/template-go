package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"template-go/internal/delivery/http/routes"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Attach root route
	r.Mount("/", routes.RootRoutes())

	return r
}
