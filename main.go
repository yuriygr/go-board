package main

import (
	"context"
	"encoding/gob"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
)

// APIVersion1 - Const for api versions
const APIVersion1 = "v1"

func main() {
	gob.Register(User{})

	session := NewSession()
	storage := NewStorage()
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Set-Cookie"},
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

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, ErrMethodNotAllowed())
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, ErrMethodNotAllowed())
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(AuthCtx(session))
		r.Use(APIVersionCtx(APIVersion1))
		r.Mount("/boards", boardsResource{storage, session}.Routes())
		r.Mount("/topics", topicsResource{storage, session}.Routes())
		r.Mount("/pages", pagesResource{storage, session}.Routes())
		r.Mount("/bugs", bugsResource{storage, session}.Routes())
		r.Mount("/users", usersResource{storage, session}.Routes())
		r.Mount("/uploader", uploadResource{storage, session}.Routes())
	})

	http.ListenAndServe(":3000", r)
}

// AuthCtxKey - Key for context
type AuthCtxKey struct{}

// AuthCtx - Контекст с авторизацией
// Очень помогает жить и вообще
func AuthCtx(session *Session) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionNew, _ := session.Auth(r)

			if _, ok := sessionNew.Values["auth"].(bool); ok {
				sessionResponse := &SessionResponse{}
				sessionResponse.Bind(sessionNew)

				ctx := context.WithValue(r.Context(), AuthCtxKey{}, sessionResponse)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := context.WithValue(r.Context(), AuthCtxKey{}, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIVersionCtxKey - Key for context
type APIVersionCtxKey struct{}

// APIVersionCtx - Context API version
func APIVersionCtx(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), APIVersionCtxKey{}, version)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
