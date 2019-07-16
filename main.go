package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	session := NewSession()
	storage := NewStorage()
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Logger)
	r.Use(cors.Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, &SuccessResponse{
			HTTPStatusCode: 200,
			StatusText:     os.Getenv("HELLO_MESSAGE"),
		})
		return
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(APIVersionCtx("v1"))
		r.Mount("/boards", boardsResource{storage, session}.Routes())
		r.Mount("/topics", topicsResource{storage, session}.Routes())
		r.Mount("/pages", pagesResource{storage, session}.Routes())
		r.Mount("/bugs", bugsResource{storage, session}.Routes())
		r.Mount("/users", usersResource{storage, session}.Routes())
	})

	http.ListenAndServe(":3000", r)
}

// APIVersionCtxKey - Key for context
type APIVersionCtxKey struct{}

// APIVersionCtx - Context API version
func APIVersionCtx(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), APIVersionCtxKey{}, version))
			next.ServeHTTP(w, r)
		})
	}
}
