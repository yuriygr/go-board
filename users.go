package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
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

	r.Post("/login", rs.UserLogin)
	r.Post("/create", rs.UserCreate)

	return r
}

//--
// Handler methods
//--

// UserLogin - Login user
func (rs *usersResource) UserLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(AuthCtxKey{}).(*SessionResponse); ok {
		err := errors.New("You are already logged in, where will you go again?")
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	username := r.FormValue("username")
	user, err := rs.storage.GetUserByUsername(username)
	if err != nil {
		//err := errors.New("This account does not exist")
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
	sessionNew.Values["user_id"] = user.ID
	sessionNew.Values["username"] = user.Username
	sessionNew.Values["screenname"] = user.Profile.ScreenName
	sessionNew.Values["auth"] = true
	sessionNew.Save(r, w)

	sessionResponse := &SessionResponse{}
	sessionResponse.Bind(sessionNew)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 200,
		StatusText:     "Ok!",
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
	sessionNew.Values["user_id"] = user.ID
	sessionNew.Values["username"] = user.Username
	sessionNew.Values["screenname"] = user.Profile.ScreenName
	sessionNew.Values["auth"] = true
	sessionNew.Save(r, w)

	sessionResponse := &SessionResponse{}
	sessionResponse.Bind(sessionNew)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "User created, let's go!",
		Payload:        sessionResponse,
	})
}

//--
// Struct
//--

// User sructure
type User struct {
	ID        int32  `json:"-" db:"u.id"`
	Username  string `json:"username" db:"u.username"`
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

	password, err := utils.HashPassword(r.FormValue("password"))
	if err != nil {
		return errors.New("Password to fucking shitty wtf")
	}

	u.Password = password
	u.Username = r.FormValue("username")
	u.CreatedAt = time.Now().Unix()
	u.States.IsBanned = false
	u.States.IsDeleted = false
	return nil
}

// SessionResponse - Состояние юзера
type SessionResponse struct {
	UserID     int32  `json:"id"`
	Username   string `json:"username"`
	ScreenName string `json:"screenname"`
	Auth       bool   `json:"auth"`
}

// Bind - Bind structure with session
func (sr *SessionResponse) Bind(session *sessions.Session) {
	sr.UserID = session.Values["user_id"].(int32)
	sr.Username = session.Values["username"].(string)
	sr.ScreenName = session.Values["screenname"].(string)
	sr.Auth = session.Values["auth"].(bool)
}
