package main

import (
	"net/http"

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

// UserState - Состояние юзера
type UserState struct {
	UserID int
	Auth   bool
}

// UserLogin - Создание пользователя
func (rs *usersResource) UserLogin(w http.ResponseWriter, r *http.Request) {

	sessionNew, _ := rs.session.Auth(r)
	sessionNew.Values["user_id"] = 1
	sessionNew.Values["auth"] = true
	sessionNew.Save(r, w)

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Created",
		Payload: &UserState{
			UserID: sessionNew.Values["user_id"].(int),
			Auth:   sessionNew.Values["auth"].(bool),
		},
	})
}

// UserCreate - Создание пользователя
func (rs *usersResource) UserCreate(w http.ResponseWriter, r *http.Request) {
	request := &UserCreateRequest{}
	if err := request.Bind(r); err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	user, err := rs.storage.CreateUser(request)
	if err != nil {
		render.Render(w, r, ErrBadRequest(err))
		return
	}

	/*sessionNew, _ := rs.session.Auth(r)
	userState := &UserState{}

	if sessionNew.Values["user_id"] != nil {
		userState.UserID = sessionNew.Values["user_id"].(int)
	}

	if sessionNew.Values["auth"] != nil {
		userState.Auth = sessionNew.Values["auth"].(bool)
	}*/

	render.Render(w, r, &SuccessResponse{
		HTTPStatusCode: 201,
		StatusText:     "Created",
		Payload:        user,
	})
}

//--
// Struct
//--

// User sructure
type User struct {
	ID        int64  `db:"u.id"`
	Username  string `db:"u.username"`
	Password  string `db:"u.password"`
	Email     string `db:"u.email"`
	CreatedAt int64  `db:"u.created_at"`
	States    struct {
		IsBanned  int8 `db:"u.is_banned"`
		IsDeleted int8 `db:"u.is_deleted"`
	} `db:""`
}
