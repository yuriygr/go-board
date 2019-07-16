package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type bugsResource struct {
	storage *Storage
	session *Session
}

func (rs bugsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.BugsList)
	r.Post("/", rs.BugsCreate)

	return r
}

//--
// Handler methods
//--

// BugsList - Return list of Bugs
func (rs *bugsResource) BugsList(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, &ErrResponse{
		HTTPStatusCode: 405,
		StatusText:     "Method not allowed",
	})
}

// BugsCreate - Handler
func (rs *bugsResource) BugsCreate(w http.ResponseWriter, r *http.Request) {
	request := &BugCreateRequest{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	bug, err := rs.storage.CreateBugReport(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// TODO: Notification
	// go rs.notify.Send(bug)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "The bug report was created successfully.",
		Payload:        bug,
	})
}

//--
// Struct
//--

// Bug structure
type Bug struct {
	Number int `json:"number"`
}
