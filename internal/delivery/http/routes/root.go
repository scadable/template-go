package routes

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"template-go/pkg/logger"
)

func RootRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", helloWorld)
	return r
}

// @Summary Hello World endpoint
// @Tags Root
// @Produce plain
// @Success 200 {string} string "Hello, World!"
// @Router / [get]
func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("Hello, World!"))
	if err != nil {
		logger.Error(r.Context(), "failed to write response", zap.Error(err))
	}

}
