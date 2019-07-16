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

// TopicCreateRequest - Request for submit
type TopicCreateRequest struct {
	Type           string `db:"type"`
	BoardID        int    `db:"board_id"`
	Subject        string `db:"subject"`
	Message        string `db:"message"`
	CreatedAt      int64  `db:"created_at"`
	BumpedAt       int64  `db:"bumped_at"`
	UserIP         string `db:"user_ip"`
	IsClosed       int8   `db:"is_closed"`
	IsPinned       int8   `db:"is_pinned"`
	IsDeleted      int8   `db:"is_deleted"`
	AllowAttach    int8   `db:"allow_attach"`
	CommentsClosed int8   `db:"comments_closed"`
}

// Bind - Bind HTTP request data and validate it
func (p *TopicCreateRequest) Bind(r *http.Request) error {
	// Bind
	p.Type = "normal"
	p.BoardID = 7
	p.Subject = r.FormValue("subject")
	p.Message = r.FormValue("message")
	p.CreatedAt = time.Now().Unix()
	p.BumpedAt = time.Now().Unix()
	p.UserIP = r.RemoteAddr
	p.IsClosed = 0
	p.IsPinned = 0
	p.IsDeleted = 0
	p.AllowAttach = 1
	p.CommentsClosed = 0

	// Validate
	if p.Subject == "" {
		return errors.New("Subject must be filled")
	}
	if p.Message == "" {
		return errors.New("Message must be filled")
	}
	if len(p.Message) < 15 {
		return errors.New("Message is too short")
	}

	return nil
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
func (s *Storage) CreateTopic(request *TopicCreateRequest) (*Topic, error) {
	result, err := s.db.NamedExec(`INSERT INTO topics (type, board_id, subject, message, created_at, bumped_at, user_ip, is_closed, is_pinned, is_deleted, allow_attach, comments_closed) VALUES (:type, :board_id, :subject, :message, :created_at, :bumped_at, :user_ip, :is_closed, :is_pinned, :is_deleted, :allow_attach, :comments_closed)`, request)
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
// Comments methods & structs
// --

// CommentCreateRequest - Request for submit
type CommentCreateRequest struct {
	TopicID   int
	Message   string
	CreatedAt int64
	UserIP    string
}

// Bind - Bind HTTP request data and validate it
func (p *CommentCreateRequest) Bind(r *http.Request) error {
	// Bind
	p.TopicID = 1
	p.Message = r.FormValue("message")
	p.CreatedAt = time.Now().Unix()
	p.UserIP = r.RemoteAddr

	// Validate
	if p.Message == "" {
		return errors.New("Message must be filled")
	}
	if len(p.Message) < 15 {
		return errors.New("Message is too short")
	}
	return nil
}

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

// CreateComment - Create comemnt with data
func (s *Storage) CreateComment(request *CommentCreateRequest) (*Comment, error) {
	fmt.Println(request)
	return &Comment{}, nil
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
	// Bind
	p.IP = r.RemoteAddr
	p.Description = r.FormValue("description")
	p.Email = r.FormValue("email")
	p.CreatedAt = time.Now().Unix()

	// Validate
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
