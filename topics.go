package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type topicsResource struct {
	storage *Storage
	session *Session
}

func (rs topicsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(rs.PaginationCtx).Get("/", rs.TopicsList)
	r.Post("/", rs.TopicCreate)

	r.Route("/{topicID:[0-9]+}", func(r chi.Router) {
		r.With(rs.TopicCtx).Get("/", rs.TopicGet)
		r.With(rs.CommentsCtx).Get("/comments", rs.TopicCommentsGet)
		r.Post("/comments", rs.CommentCreate)
		r.Post("/report", rs.ReportCreate)
	})

	return r
}

//--
// Middleware
//--

// TopicsCtxKey - Key for context
type TopicsCtxKey struct{}

// PaginationCtx middleware для пагинации топиков
func (rs *topicsResource) PaginationCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := &TopicsRequest{Sort: "bumped_at", Page: 1, Limit: 30}
		topics, err := rs.storage.GetTopicsList(request)
		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), TopicsCtxKey{}, topics)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TopicCtxKey - Key for context
type TopicCtxKey struct{}

// TopicCtx middleware is used to load an Topic object from
// the URL parameters passed through as the request. In case
// the Topic could not be found, we stop here and return a 404.
func (rs *topicsResource) TopicCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var topic *Topic
		var err error

		if topicID := chi.URLParam(r, "topicID"); topicID != "" {
			topicID, _ := strconv.ParseInt(topicID, 10, 64)
			topic, err = rs.storage.GetTopicByID(topicID)
		} else {
			render.Render(w, r, ErrBadRequest(errors.New("ID needed")))
			return
		}

		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), TopicCtxKey{}, topic)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CommentsCtxKey middleware для вывода комментариев топика
type CommentsCtxKey struct{}

// CommentsCtx
func (rs *topicsResource) CommentsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var comments []*Comment
		var err error

		if topicID := chi.URLParam(r, "topicID"); topicID != "" {
			topicID, _ := strconv.Atoi(topicID)
			comments, err = rs.storage.GetCommentsList(topicID)
		} else {
			render.Render(w, r, ErrBadRequest(errors.New("ID needed")))
			return
		}

		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), CommentsCtxKey{}, comments)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//--
// Handler methods
//--

// TopicsList - Вывод списка топиков
func (rs *topicsResource) TopicsList(w http.ResponseWriter, r *http.Request) {
	topics := r.Context().Value(TopicsCtxKey{}).([]*Topic)

	if err := render.RenderList(w, r, NewTopicsListResponse(topics)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TopicGet - Вывод объекта топика
func (rs *topicsResource) TopicGet(w http.ResponseWriter, r *http.Request) {
	topic := r.Context().Value(TopicCtxKey{}).(*Topic)

	if err := render.Render(w, r, topic); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TopicCreate - Создает объект топика
func (rs *topicsResource) TopicCreate(w http.ResponseWriter, r *http.Request) {
	request := &Topic{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	topic, err := rs.storage.CreateTopic(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	if err := topic.Render(w, r); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Topic created!",
		Payload:        topic,
	})
}

// TopicCommentsGet - List of comments on topic
func (rs *topicsResource) TopicCommentsGet(w http.ResponseWriter, r *http.Request) {
	comments := r.Context().Value(CommentsCtxKey{}).([]*Comment)

	if err := render.RenderList(w, r, NewCommentListResponse(comments)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CommentCreate - Создает объект комментария
func (rs *topicsResource) CommentCreate(w http.ResponseWriter, r *http.Request) {
	request := &Comment{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	comment, err := rs.storage.CreateComment(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	if err := comment.Render(w, r); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Comment created!",
		Payload:        comment,
	})
}

// ReportCreate - Создает жалобы на топик
func (rs *topicsResource) ReportCreate(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Report created!",
	})
}

//--
// Struct
//--

// Topic structure
type Topic struct {
	ID            int    `json:"id" db:"t.id"`
	Type          string `json:"type" db:"t.type"`
	BoardID       int    `json:"-" db:"t.board_id"`
	Name          string `json:"name" db:"-"`
	Subject       string `json:"subject" db:"t.subject"`
	Message       string `json:"message" db:"t.message"`
	CreatedAt     int64  `json:"created_at" db:"t.created_at"`
	BumpedAt      int64  `json:"bumped_at" db:"t.bumped_at"`
	UserIP        string `json:"-" db:"t.user_ip"`
	CommentsCount int    `json:"comments_count" db:"comments_count"`
	States        struct {
		IsClosed  int8 `json:"is_closed" db:"t.is_closed"`
		IsPinned  int8 `json:"is_pinned" db:"t.is_pinned"`
		IsDeleted int8 `json:"-" db:"t.is_deleted"`
	} `json:"states" db:""`
	Options struct {
		AllowAttach    int8 `json:"allow_attach" db:"t.allow_attach"`
		CommentsClosed int8 `json:"comments_closed" db:"t.comments_closed"`
	} `json:"options" db:""`
	Board struct {
		Title string `json:"title" db:"b.title"`
		Slug  string `json:"slug" db:"b.slug"`
	} `json:"board" db:""`

	Attachments []interface{} `json:"attachments" db:"-"`
}

// Render - Render, wtf
func (t *Topic) Render(w http.ResponseWriter, r *http.Request) error {
	t.Name = "Anonüümne"
	t.Attachments = make([]interface{}, 0)
	return nil
}

// Bind - Bind HTTP request data and validate it
func (t *Topic) Bind(r *http.Request) error {
	// Bind
	t.Type = "normal"
	t.BoardID = 7
	t.Subject = r.FormValue("subject")
	t.Message = r.FormValue("message")
	t.CreatedAt = time.Now().Unix()
	t.BumpedAt = time.Now().Unix()
	t.UserIP = r.RemoteAddr
	t.States.IsClosed = 0
	t.States.IsPinned = 0
	t.States.IsDeleted = 0
	t.Options.AllowAttach = 1
	t.Options.CommentsClosed = 0

	// Validate
	if t.Subject == "" {
		return errors.New("Subject must be filled")
	}
	if t.Message == "" {
		return errors.New("Message must be filled")
	}
	if len(t.Message) < 15 {
		return errors.New("Message is too short")
	}

	return nil
}

// NewTopicsListResponse -
func NewTopicsListResponse(topics []*Topic) []render.Renderer {
	list := []render.Renderer{}
	for _, topic := range topics {
		list = append(list, topic)
	}
	return list
}

// Comment structure
type Comment struct {
	ID        int    `json:"id" db:"c.id"`
	TopicID   int    `json:"topic_id" db:"c.topic_id"`
	Message   string `json:"message" db:"c.message"`
	CreatedAt int64  `json:"created_at" db:"c.created_at"`
	UserIP    string `json:"-" db:"c.user_ip"`
	States    struct {
		IsPinned  int8 `json:"is_pinned" db:"c.is_pinned"`
		IsDeleted int8 `json:"is_deleted" db:"c.is_deleted"`
	} `json:"states" db:""`
}

// Render - Render, wtf
func (c *Comment) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Bind - Bind HTTP request data and validate it
func (c *Comment) Bind(r *http.Request) error {
	c.Message = r.FormValue("message")
	c.CreatedAt = time.Now().Unix()
	c.UserIP = r.RemoteAddr
	c.States.IsPinned = 0
	c.States.IsDeleted = 0

	if topicID := chi.URLParam(r, "topicID"); topicID != "" {
		topicID, _ := strconv.Atoi(topicID)
		c.TopicID = topicID
	} else {
		return errors.New("You must specify the number of the topic")
	}

	if c.Message == "" {
		return errors.New("Message must be filled")
	}
	if len(c.Message) < 15 {
		return errors.New("Message is too short")
	}
	return nil
}

// NewCommentListResponse -
func NewCommentListResponse(comments []*Comment) []render.Renderer {
	list := []render.Renderer{}
	for _, comment := range comments {
		list = append(list, comment)
	}
	return list
}
