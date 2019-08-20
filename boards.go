package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type boardsResource struct {
	storage *Storage
	session *Session
}

func (rs boardsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(rs.BoardsCtx).Get("/", rs.BoardsList)

	return r
}

//--
// Middleware
//--

// BoardsCtxKey - Key for context
type BoardsCtxKey struct{}

// BoardsCtx middleware для запроса к...
func (rs *boardsResource) BoardsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		boards, err := rs.storage.GetBoardsList()
		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), BoardsCtxKey{}, boards)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//--
// Handler methods
//--

// BoardsList - Return list of Boards
func (rs *boardsResource) BoardsList(w http.ResponseWriter, r *http.Request) {
	boards := r.Context().Value(BoardsCtxKey{}).([]*Board)

	if err := render.RenderList(w, r, NewBoardsListResponse(boards)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

//--
// Struct
//--

// Board structure
type Board struct {
	ID        int64  `json:"-" db:"b.id"`
	Title     string `json:"title" db:"b.title"`
	Slug      string `json:"slug" db:"b.slug"`
	Type      string `json:"type" db:"b.type"`
	Available bool   `json:"-" db:"b.available"`
	NSFW      bool   `json:"nsfw" db:"b.nsfw"`
}

// Render - Render, wtf
func (b *Board) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// NewBoardsListResponse -
func NewBoardsListResponse(boards []*Board) []render.Renderer {
	list := []render.Renderer{}
	for _, board := range boards {
		list = append(list, board)
	}
	return list
}
