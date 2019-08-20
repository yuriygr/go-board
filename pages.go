package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type pagesResource struct {
	storage *Storage
	session *Session
}

func (rs pagesResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.PagesList)
	r.Route("/{pageSlug}", func(r chi.Router) {
		r.Use(rs.PageCtx)
		r.Get("/", rs.PageGet)
	})

	return r
}

//--
// Middleware
//--

// PageCtxKey - Key for context
type PageCtxKey struct{}

// PageCtx middleware is used to load an Page object from
// the URL parameters passed through as the request. In case
// the Page could not be found, we stop here and return a 404.
func (rs *pagesResource) PageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var page *Page
		var err error

		if pageSlug := chi.URLParam(r, "pageSlug"); pageSlug != "" {
			page, err = rs.storage.GetPageBySlug(pageSlug)
		} else {
			render.Render(w, r, ErrBadRequest(errors.New("Slug needed")))
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), PageCtxKey{}, page)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//--
// Handler methods
//--

// PagesList - Return list of Page
func (rs *pagesResource) PagesList(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, &ErrResponse{
		HTTPStatusCode: 405,
		StatusText:     "Method not allowed",
	})
}

// PageGet - Возвращает структуру ресурса
func (rs *pagesResource) PageGet(w http.ResponseWriter, r *http.Request) {
	page := r.Context().Value(PageCtxKey{}).(*Page)

	if err := render.Render(w, r, page); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

//--
// Struct
//--

// Page struct
type Page struct {
	ID         int32  `json:"-" db:"p.id"`
	Slug       string `json:"slug" db:"p.slug"`
	Title      string `json:"title" db:"p.title"`
	Content    string `json:"content" db:"p.content"`
	CreatedAt  int64  `json:"created_at,omitempty" db:"p.created_at"`
	ModifiedIn int64  `json:"modified_in,omitempty" db:"p.modified_in"`
	States     struct {
		IsComments bool `json:"is_comments,omitempty" db:"p.is_comments"`
		IsHidden   bool `json:"-" db:"p.is_hidden"`
	} `json:"states" db:""`
	Meta struct {
		Description string `json:"meta_description" db:"p.meta_description"`
		Keywords    string `json:"meta_keywords" db:"p.meta_keywords"`
	} `json:"meta" db:""`
}

// Render - render, wtf
func (p *Page) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
