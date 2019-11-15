package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/yuriygr/go-board/utils"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type usersResource struct {
	storage *Storage
	session *Session
}

func (rs usersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{userID}", func(r chi.Router) {
		r.Use(rs.UserCtx)
		r.Get("/", rs.UserGet)
		r.Get("/statistic", rs.StatisticGet)
	})
	r.Post("/login", rs.UserLogin)
	r.Post("/create", rs.UserCreate)
	r.Get("/session", rs.UserSession)

	return r
}

//--
// Middleware
//--

// UserCtxKey - Key for context
type UserCtxKey struct{}

// UserCtx middleware is used to load an User object from
// the URL parameters passed through as the request. In case
// the User could not be found, we stop here and return a 404.
func (rs *usersResource) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *User
		var err error

		if userID := chi.URLParam(r, "userID"); userID != "" {
			userID, _ := strconv.ParseInt(userID, 10, 64)
			user, err = rs.storage.GetUserByID(userID)
		} else {
			render.Render(w, r, ErrBadRequest(errors.New("User ID needed")))
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound(err))
			return
		}

		ctx := context.WithValue(r.Context(), UserCtxKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//--
// Handler methods
//--

// UserGet - Get user
func (rs *usersResource) UserGet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey{}).(*User)

	if err := render.Render(w, r, user); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// StatisticGet * Get user statistic
func (rs *usersResource) StatisticGet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey{}).(*User)

	statistic, err := rs.storage.GetUserStatistic(user.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	if err := render.Render(w, r, statistic); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UserLogin - Login user
func (rs *usersResource) UserLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); ok {
		err := errors.New("You are already logged in, where will you go again?")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	username := r.FormValue("username")
	username = utils.EscapeString(username)
	user, err := rs.storage.GetUserByUsername(username)
	if err != nil {
		err := errors.New("Invalid username or password")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	password := r.FormValue("password")
	if !utils.CheckPasswordHash(password, user.Password) {
		err := errors.New("Invalid username or password")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	if user.States.IsBanned {
		err := errors.New("Sorry Mario, the Princess is in another castle")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	if user.States.IsDeleted {
		err := errors.New("This account does not exist")
		render.Render(w, r, ErrBadRequest(err))
		return
	}
	// Ok, user exist and password correct...
	// Let's create session!

	sessionNew, _ := rs.session.Auth(r)
	sessionNew.Values["user"] = *user
	sessionNew.Values["auth"] = true
	if err := sessionNew.Save(r, w); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	sessionResponse := &SessionResponse{}
	sessionResponse.Bind(sessionNew)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 200,
		StatusText:     "You are successfully logged in.",
		Payload:        sessionResponse,
	})
}

// UserCreate - Создание пользователя
func (rs *usersResource) UserCreate(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); ok {
		err := errors.New("You are already logged in, where will you go again?")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	request := &User{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	user, err := rs.storage.CreateUser(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	// Ok, user created successfully...
	// Let's create session!

	sessionNew, _ := rs.session.Auth(r)
	sessionNew.Values["user"] = *user
	sessionNew.Values["auth"] = true
	if err := sessionNew.Save(r, w); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	sessionResponse := &SessionResponse{}
	sessionResponse.Bind(sessionNew)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Account successfully created, let's go!",
		Payload:        sessionResponse,
	})
}

// UserSession - Check session
func (rs *usersResource) UserSession(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); !ok {
		err := errors.New("You are not authorized, what session is it for you?")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	sessionNew, _ := rs.session.Auth(r)
	sessionResponse := &SessionResponse{}
	sessionResponse.Bind(sessionNew)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Your session",
		Payload:        sessionResponse,
	})
}

//--
// Struct
//--

// User sructure
type User struct {
	ID        int64  `json:"id" db:"u.id"`
	Username  string `json:"-" db:"u.username"`
	Password  string `json:"-" db:"u.password"`
	CreatedAt int64  `json:"-" db:"u.created_at"`
	Profile   struct {
		ScreenName string `json:"screen_name" db:"up.screen_name"`
		Sex        string `json:"sex" db:"up.sex"`
	} `json:"profile" db:""`
	States struct {
		IsBanned  bool `json:"is_banned" db:"u.is_banned"`
		IsDeleted bool `json:"is_deleted" db:"u.is_deleted"`
	} `json:"states" db:""`
}

// Render - Render, wtf
func (u *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Bind - Bind HTTP request data and validate it
func (u *User) Bind(r *http.Request) error {

	if r.FormValue("username") == "" {
		return errors.New("Username must be filled")
	}
	if r.FormValue("password") == "" {
		return errors.New("Password must be filled")
	}
	if r.FormValue("password") != r.FormValue("password_confirm") {
		return errors.New("Password confirmation does not match the password")
	}

	username := utils.EscapeString(r.FormValue("username"))

	password, err := utils.HashPassword(r.FormValue("password"))
	if err != nil {
		return errors.New("Password to fucking shitty wtf")
	}

	u.Password = password
	u.Username = username
	u.CreatedAt = time.Now().Unix()
	u.Profile.ScreenName = username
	u.States.IsBanned = false
	u.States.IsDeleted = false
	return nil
}

// UserStatistic - User stats
type UserStatistic struct {
	ID        int64 `json:"-" db:"us.id"`
	UserID    int64 `json:"user_id" db:"us.user_id"`
	Statistic struct {
		AccountCreated  int64 `json:"account_created" db:"u.created_at"`
		СreatedTopics   int64 `json:"created_topics" db:"us.created_topics"`
		СreatedComments int64 `json:"created_comments" db:"us.created_comments"`
		UploadedFiles   int64 `json:"uploaded_files" db:"us.uploaded_files"`
	} `json:"statistic" db:""`
}

// Render - Render, wtf
func (us *UserStatistic) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
