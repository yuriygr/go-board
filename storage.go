package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	selectBoards   = "select b.* from boards as b"
	selectPages    = "select p.* from pages as p"
	selectTopics   = "select t.*, b.title, b.slug, COUNT(c.id) as comments_count from topics as t left join boards as b on t.board_id = b.id left join comments as c on c.topic_id = t.id"
	selectComments = "select c.* from comments as c"

	selectPageBySlug        = selectPages + " where p.slug = '%s'"
	selectTopicsByID        = selectTopics + " where t.id = '%d' group by t.id"
	selectTopicsPaginated   = selectTopics + " group by t.id order by t.is_pinned desc, %s desc limit %d"
	selectCommentByID       = selectComments + " where c.id = '%d'"
	selectCommentsByTopicID = selectComments + " where c.topic_id = '%d' order by c.is_pinned desc, c.created_at asc"
)

// NewStorage - init new storage
func NewStorage() *Storage {
	db, err := sqlx.Connect("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Hour)
	return &Storage{db}
}

// Storage - Нечто такое обстрактное я хз
// попозже придумаю что тут написать так то
// штука крутая.
type Storage struct {
	db *sqlx.DB
}

//--
// Boards methods
//--

// GetBoardsList - Get list of boards
func (s *Storage) GetBoardsList() ([]*Board, error) {
	boards := []*Board{}

	err := s.db.Select(&boards, selectBoards)
	if err != nil {
		return nil, err
	}

	return boards, nil
}

//--
// Page methods
//--

// GetPageBySlug - Возвращает страницу по Slug
func (s *Storage) GetPageBySlug(slug string) (*Page, error) {
	page := Page{}
	sql := fmt.Sprintf(selectPageBySlug, slug)

	err := s.db.Get(&page, sql)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

//--
// Topic methods & structs
//--

// TopicsRequest - Request for fetch topics
type TopicsRequest struct {
	Sort  string
	Page  int16
	Limit int8
}

// GetTopicsList - Получает список топиков с параметрами
func (s *Storage) GetTopicsList(request *TopicsRequest) ([]*Topic, error) {
	topics := []*Topic{}
	sql := fmt.Sprintf(selectTopicsPaginated, request.Sort, request.Limit)

	err := s.db.Select(&topics, sql)
	if err != nil {
		return nil, err
	}

	return topics, nil
}

// GetTopicByID - Возвращает топик по Id
func (s *Storage) GetTopicByID(id int64) (*Topic, error) {
	topic := Topic{}
	sql := fmt.Sprintf(selectTopicsByID, id)

	err := s.db.Get(&topic, sql)
	if err != nil {
		return nil, err
	}

	return &topic, nil
}

// CreateTopic - Create topic with data
func (s *Storage) CreateTopic(request *Topic) (*Topic, error) {
	result, err := s.db.NamedExec(`INSERT INTO topics (type, board_id, subject, message, created_at, bumped_at, user_ip, is_closed, is_pinned, is_deleted, allow_attach, comments_closed) VALUES (:t.type, :t.board_id, :t.subject, :t.message, :t.created_at, :t.bumped_at, :t.user_ip, :t.is_closed, :t.is_pinned, :t.is_deleted, :t.allow_attach, :t.comments_closed)`, request)
	if err != nil {
		return nil, err
	}

	topicID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	topic, err := s.GetTopicByID(topicID)
	if err != nil {
		return nil, err
	}

	return topic, nil
}

// --
// Comments methods
// --

// GetCommentsList -
func (s *Storage) GetCommentsList(id int) ([]*Comment, error) {
	comments := []*Comment{}
	sql := fmt.Sprintf(selectCommentsByTopicID, id)

	err := s.db.Select(&comments, sql)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// GetCommentByID - Возвращает топик по Id
func (s *Storage) GetCommentByID(id int64) (*Comment, error) {
	comment := Comment{}
	sql := fmt.Sprintf(selectCommentByID, id)

	err := s.db.Get(&comment, sql)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// CreateComment - Create comemnt with data
func (s *Storage) CreateComment(request *Comment) (*Comment, error) {
	result, err := s.db.NamedExec(`INSERT INTO comments (topic_id, message, created_at, user_ip, is_pinned, is_deleted) VALUES (:c.topic_id, :c.message, :c.created_at, :c.user_ip, :c.is_pinned, :c.is_deleted)`, request)
	if err != nil {
		return nil, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	comment, err := s.GetCommentByID(commentID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

//--
// Bugs methods & structs
//--

// BugCreateRequest - Request for create bug report method
type BugCreateRequest struct {
	IP                 string
	Description, Email string
	CreatedAt          int64
}

// Bind - Bind HTTP request data and validate it
func (p *BugCreateRequest) Bind(r *http.Request) error {
	p.IP = r.RemoteAddr
	p.Description = r.FormValue("description")
	p.Email = r.FormValue("email")
	p.CreatedAt = time.Now().Unix()

	if p.Description == "" {
		return errors.New("Description must be filled")
	}
	if len(p.Description) < 15 {
		return errors.New("Description is too short")
	}

	return nil
}

// GetBugByID - Возвращает Bug по Id
func (s *Storage) GetBugByID(id int) (*Bug, error) {
	return &Bug{Number: 1}, errors.New("Bug not found")
}

// CreateBugReport - Create bug report with data
func (s *Storage) CreateBugReport(request *BugCreateRequest) (*Bug, error) {
	fmt.Println(request)
	return &Bug{Number: 1}, nil
}

//--
// Users methods & structs
//--

// UserCreateRequest - Request for create user
type UserCreateRequest struct {
	Login     string
	Password  string
	CreatedAt int64
}

// Bind - Bind HTTP request data and validate it
func (p *UserCreateRequest) Bind(r *http.Request) error {
	p.Login = r.FormValue("login")
	p.Password = r.FormValue("password")
	p.CreatedAt = time.Now().Unix()

	if p.Login == "" {
		return errors.New("Login must be filled")
	}
	if p.Password == "" {
		return errors.New("Password must be filled")
	}
	return nil
}

func (s *Storage) CreateUser(request *UserCreateRequest) (*User, error) {
	return &User{}, nil
}
