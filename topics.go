package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/yuriygr/go-board/utils"
)

type topicsResource struct {
	storage *Storage
	session *Session
}

func (rs topicsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.With(rs.PaginationCtx).Get("/", rs.TopicsList)
		r.Post("/", rs.TopicCreate)
	})

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
		request := &TopicsRequest{Sort: "bumped_at", Page: 1, Limit: 30} // Initial state
		if err := request.Bind(r); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		topics, err := rs.storage.GetTopicsList(request)
		if err != nil {
			render.Render(w, r, ErrBadRequest(err))
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
			err := errors.New("ID needed")
			render.Render(w, r, ErrBadRequest(err))
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
		request := &CommentsRequest{} // Initial state
		if err := request.Bind(r); err != nil {
			render.Render(w, r, ErrBadRequest(err))
			return
		}

		comments, err := rs.storage.GetCommentsList(request)
		if err != nil {
			render.Render(w, r, ErrBadRequest(err))
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
// Тут же задается значение профиля по умолчанию
// А так же валидируется борда, куда будет добавлен пост
func (rs *topicsResource) TopicCreate(w http.ResponseWriter, r *http.Request) {
	request := &Topic{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// Before we started, check board
	board, err := rs.storage.GetBoardBySlug(r.FormValue("board"))
	if err != nil {
		err := errors.New("Board not found")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// Set board
	request.BoardID = board.ID

	// Now, we associate the user with the topic
	if auth, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); ok {
		request.UserID = auth.User.ID
	} else {
		request.UserID = 1 // Default Anon profile
	}

	topic, err := rs.storage.CreateTopic(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// Tracking user stats
	go rs.storage.UpdateUserStatistic(topic.UserID, "created_topics")

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

// CommentCreate - Creates a comment object
// There is also a check for the presence of a thread and its statuses.
func (rs *topicsResource) CommentCreate(w http.ResponseWriter, r *http.Request) {
	request := &Comment{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// Before, check is topic exist
	topic, err := rs.storage.GetTopicByID(request.TopicID)
	if err != nil {
		err := errors.New("Topic not exist")
		render.Render(w, r, ErrForbidden(err))
		return
	}

	// And if exist, closed it, or not
	if topic.States.IsClosed {
		err := errors.New("Topic closed")
		render.Render(w, r, ErrForbidden(err))
		return
	}

	// And deleted it
	if topic.States.IsDeleted {
		err := errors.New("Topic not exist")
		render.Render(w, r, ErrForbidden(err))
		return
	}

	// Now, we associate the user with the comment
	if auth, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); ok {
		request.UserID = auth.User.ID
	} else {
		request.UserID = 1 // Default Anon profile
	}

	comment, err := rs.storage.CreateComment(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// And then, bump topic
	go rs.storage.UpdateTopicBumpTime(comment)

	// Tracking user stats
	go rs.storage.UpdateUserStatistic(comment.UserID, "created_comments")

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
// Topic structure
//--

// Topic structure
type Topic struct {
	ID            int64  `json:"id" db:"t.id"`
	Type          string `json:"type" db:"t.type"`
	BoardID       int64  `json:"-" db:"t.board_id"`
	UserID        int64  `json:"-" db:"t.user_id"`
	Subject       string `json:"subject" db:"t.subject"`
	Message       string `json:"message" db:"t.message"`
	CreatedAt     int64  `json:"created_at" db:"t.created_at"`
	BumpedAt      int64  `json:"bumped_at" db:"t.bumped_at"`
	UserIP        string `json:"-" db:"t.user_ip"`
	UserAgent     string `json:"-" db:"t.user_agent"`
	CommentsCount int    `json:"comments_count" db:"comments_count"`
	User          struct {
		ScreenName string `json:"screen_name" db:"up.screen_name"`
		IsAdmin    bool   `json:"is_admin" db:"-"`
	} `json:"user" db:""`
	Board struct {
		Title string `json:"title" db:"b.title"`
		Slug  string `json:"slug" db:"b.slug"`
	} `json:"board" db:""`
	States struct {
		IsClosed  bool `json:"is_closed" db:"t.is_closed"`
		IsPinned  bool `json:"is_pinned" db:"t.is_pinned"`
		IsDeleted bool `json:"-" db:"t.is_deleted"`
	} `json:"states" db:""`
	Options struct {
		AllowAttach     bool `json:"allow_attach" db:"t.allow_attach"`
		OnlyAnonymously bool `json:"only_anonymously" db:"t.only_anonymously"`
	} `json:"options" db:""`

	Attachments []interface{} `json:"attachments" db:"-"`
}

// Render - Render, wtf
func (t *Topic) Render(w http.ResponseWriter, r *http.Request) error {
	t.User.IsAdmin = false
	t.Attachments = make([]interface{}, 0)
	return nil
}

// Bind - Bind HTTP request data and validate it
func (t *Topic) Bind(r *http.Request) error {
	if r.FormValue("board") == "" {
		return errors.New("Board must be filled")
	}
	if r.FormValue("subject") == "" {
		return errors.New("Subject must be filled")
	}
	if r.FormValue("message") == "" {
		return errors.New("Message must be filled")
	}
	if len(r.FormValue("message")) < 15 {
		return errors.New("Message is too short")
	}

	// Awesome parser for markup

	message, err := utils.ParseMessage(r.FormValue("message"))
	if err != nil {
		return errors.New("Message so borred")
	}

	t.Type = "normal"
	t.Subject = r.FormValue("subject")
	t.Message = message
	t.CreatedAt = time.Now().Unix()
	t.BumpedAt = time.Now().Unix()
	t.UserIP = r.Header.Get("X-FORWARDED-FOR")
	t.UserAgent = r.UserAgent()
	t.States.IsClosed = false
	t.States.IsPinned = false
	t.States.IsDeleted = false
	t.Options.AllowAttach = true
	t.Options.OnlyAnonymously = false

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

//--
// Comment structure
//--

// Comment structure
type Comment struct {
	ID        int64  `json:"id" db:"c.id"`
	TopicID   int64  `json:"topic_id" db:"c.topic_id"`
	UserID    int64  `json:"-" db:"c.user_id"`
	Message   string `json:"message" db:"c.message"`
	CreatedAt int64  `json:"created_at" db:"c.created_at"`
	UserIP    string `json:"-" db:"c.user_ip"`
	UserAgent string `json:"-" db:"c.user_agent"`
	User      struct {
		ScreenName string `json:"screen_name" db:"up.screen_name"`
		IsAdmin    bool   `json:"is_admin" db:"-"`
	} `json:"user" db:""`
	States struct {
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
	if topicID := chi.URLParam(r, "topicID"); topicID != "" {
		topicID, _ := strconv.ParseInt(topicID, 10, 64)
		c.TopicID = topicID
	} else {
		return errors.New("You must specify the number of the topic")
	}

	if r.FormValue("message") == "" {
		return errors.New("Message must be filled")
	}

	// Awesome parser for markup

	message, err := utils.ParseMessage(r.FormValue("message"))
	if err != nil {
		return errors.New("Message so borred")
	}

	c.Message = message
	c.CreatedAt = time.Now().Unix()
	c.UserIP = r.Header.Get("X-FORWARDED-FOR")
	c.UserAgent = r.UserAgent()
	c.States.IsPinned = 0
	c.States.IsDeleted = 0

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
