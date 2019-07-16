package main

import (
	"errors"
	"net/http"
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

	r.Post("/login", rs.UserLogin)
	r.Post("/create", rs.UserCreate)

	return r
}

//--
// Handler methods
//--

// UserLogin - Создание пользователя
func (rs *usersResource) UserLogin(w http.ResponseWriter, r *http.Request) {

	sessionNew, _ := rs.session.Auth(r)
	/*sessionNew.Values["user_id"] = 1
	sessionNew.Values["auth"] = true
	sessionNew.Save(r, w)*/

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 200,
		StatusText:     "Ok!",
		Payload: &SessionResponse{
			UserID: sessionNew.Values["user_id"].(int32),
			Auth:   sessionNew.Values["auth"].(bool),
		},
	})
}

// UserCreate - Создание пользователя
func (rs *usersResource) UserCreate(w http.ResponseWriter, r *http.Request) {
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
	sessionNew.Values["auth"] = true
	sessionNew.Save(r, w)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Created",
		Payload: &SessionResponse{
			UserID:   sessionNew.Values["user_id"].(int32),
			Username: sessionNew.Values["username"].(string),
			Auth:     sessionNew.Values["auth"].(bool),
		},
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
	Email     string `json:"email" db:"u.email"`
	CreatedAt int64  `json:"-" db:"u.created_at"`
	States    struct {
		IsBanned  int8 `json:"is_banned" db:"u.is_banned"`
		IsDeleted int8 `json:"is_deleted" db:"u.is_deleted"`
	} `json:"states" db:""`
}

// Render - Render, wtf
func (u *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Bind - Bind HTTP request data and validate it
func (u *User) Bind(r *http.Request) error {
	u.Username = r.FormValue("username")
	u.CreatedAt = time.Now().Unix()
	u.States.IsBanned = 0
	u.States.IsDeleted = 0

	password, err := utils.HashPassword(r.FormValue("password"))
	if err != nil {
		return errors.New("Password to fucking shitty")
	}
	u.Password = password

	if u.Username == "" {
		return errors.New("Username must be filled")
	}
	if u.Password == "" {
		return errors.New("Password must be filled")
	}
	return nil
}

// SessionResponse - Состояние юзера
type SessionResponse struct {
	UserID   int32  `json:"id"`
	Username string `json:"username"`
	Auth     bool   `json:"auth"`
}
