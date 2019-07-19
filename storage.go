package main

import (
	"errors"
	"fmt"
	"log"
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
	selectUsers    = "select u.* from users as u"

	selectPageBySlug        = selectPages + " where p.slug = '%s'"
	selectTopicsByID        = selectTopics + " where t.id = '%d' group by t.id"
	selectTopicsPaginated   = selectTopics + " group by t.id order by t.is_pinned desc, %s desc limit %d"
	selectCommentByID       = selectComments + " where c.id = '%d'"
	selectCommentsByTopicID = selectComments + " where c.topic_id = '%d' order by c.is_pinned desc, c.created_at asc"
	selectUserByID          = selectUsers + " where u.id = '%d'"
	selectUserByUsername    = selectUsers + " where u.username = '%s'"

	insertComment = "INSERT INTO comments (topic_id, message, created_at, user_ip, is_pinned, is_deleted) VALUES (:c.topic_id, :c.message, :c.created_at, :c.user_ip, :c.is_pinned, :c.is_deleted)"
	inserTopic    = "INSERT INTO topics (type, board_id, subject, message, created_at, bumped_at, user_ip, is_closed, is_pinned, is_deleted, allow_attach, comments_closed) VALUES (:t.type, :t.board_id, :t.subject, :t.message, :t.created_at, :t.bumped_at, :t.user_ip, :t.is_closed, :t.is_pinned, :t.is_deleted, :t.allow_attach, :t.comments_closed)"
	inserUser     = "INSERT INTO users (username, password, created_at, is_banned, is_deleted) VALUES (:u.username, :u.password, :u.created_at, :u.is_banned, :u.is_deleted)"
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

// GetPageBySlug - Return page by Slug
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

// GetTopicsList - Return topics list with params
func (s *Storage) GetTopicsList(request *TopicsRequest) ([]*Topic, error) {
	topics := []*Topic{}
	sql := fmt.Sprintf(selectTopicsPaginated, request.Sort, request.Limit)

	err := s.db.Select(&topics, sql)
	if err != nil {
		return nil, err
	}

	return topics, nil
}

// GetTopicByID - Return topic by ID
func (s *Storage) GetTopicByID(id int64) (*Topic, error) {
	topic := Topic{}
	sql := fmt.Sprintf(selectTopicsByID, id)

	err := s.db.Get(&topic, sql)
	if err != nil {
		return nil, err
	}

	return &topic, nil
}

// CreateTopic - Create topic and return him, or error
func (s *Storage) CreateTopic(request *Topic) (*Topic, error) {
	result, err := s.db.NamedExec(inserTopic, request)
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

// GetCommentsList - Return list of comments by topic ID
func (s *Storage) GetCommentsList(id int) ([]*Comment, error) {
	comments := []*Comment{}
	sql := fmt.Sprintf(selectCommentsByTopicID, id)

	err := s.db.Select(&comments, sql)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// GetCommentByID - Return comment by ID
func (s *Storage) GetCommentByID(id int64) (*Comment, error) {
	comment := Comment{}
	sql := fmt.Sprintf(selectCommentByID, id)

	err := s.db.Get(&comment, sql)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// CreateComment - Create comment and return him, or error
func (s *Storage) CreateComment(request *Comment) (*Comment, error) {
	result, err := s.db.NamedExec(insertComment, request)
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
// Bugs methods
//--

// GetBugByID - Return bug by ID
func (s *Storage) GetBugByID(id int) (*Bug, error) {
	return &Bug{Number: 1}, errors.New("Bug not found")
}

// CreateBugReport - Create bug report with data
func (s *Storage) CreateBugReport(request *BugCreateRequest) (*Bug, error) {
	fmt.Println(request)
	return &Bug{Number: 1}, nil
}

//--
// Users methods
//--

// GetUserByUsername - Return user by username
func (s *Storage) GetUserByUsername(username string) (*User, error) {
	user := User{}
	sql := fmt.Sprintf(selectUserByUsername, username)

	err := s.db.Get(&user, sql)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID - Return user by ID
func (s *Storage) GetUserByID(id int64) (*User, error) {
	user := User{}
	sql := fmt.Sprintf(selectUserByID, id)

	err := s.db.Get(&user, sql)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser - Create user and return him, or error
func (s *Storage) CreateUser(request *User) (*User, error) {
	result, err := s.db.NamedExec(inserUser, request)
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
