package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"gopkg.in/boj/redistore.v1"
)

var (
	key  = os.Getenv("SESSION_KEY")
	size = os.Getenv("REDIS_SIZE")
	host = os.Getenv("REDIS_HOST")
	port = os.Getenv("REDIS_PORT")
	pass = os.Getenv("REDIS_PASS")
)

// NewSession - init new cookie storage
func NewSession() *Session {
	path := fmt.Sprintf("%s:%s", host, port)
	session, err := redistore.NewRediStore(256, "tcp", path, pass, []byte(key))
	if err != nil {
		log.Fatalln(err)
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 256, // 256 Days
		HttpOnly: true,
	}

	return &Session{rs: session}
}

// Session - Нечто такое обстрактное я хз
type Session struct {
	rs *redistore.RediStore
}

// Auth - Прокси грубо говоря для стандартного Get с ключем "auth_session"
func (s *Session) Auth(r *http.Request) (*sessions.Session, error) {
	return s.rs.Get(r, "auth_session")
}

func (s *Session) Set(key string, value interface{}) error {
	return nil
}
