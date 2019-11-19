package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yuriygr/go-board/utils"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	selectBoards         = "select b.* from boards as b"
	selectPages          = "select p.* from pages as p"
	selectTopics         = "select t.*, b.title, b.slug, COUNT(c.id) as comments_count, up.user_id, up.screen_name, (select count(*) from files as f left join topics_files as tf on tf.file_id = f.id where tf.topic_id = t.id) as files_count from topics as t left join boards as b on t.board_id = b.id left join comments as c on c.topic_id = t.id left join users_profile as up on up.user_id = t.user_id"
	selectComments       = "select c.*, up.screen_name from comments as c left join users_profile as up on up.user_id = c.user_id"
	selectUsers          = "select u.*, up.screen_name, up.sex from users as u left join users_profile as up on up.user_id = u.id"
	selectUsersStatistic = "select us.*, u.created_at from users_stats as us left join users as u on us.user_id = u.id"

	selectBoardBySlug                 = selectBoards + " where b.slug = '%s'"
	selectPageBySlug                  = selectPages + " where p.slug = '%s'"
	selectCommentByID                 = selectComments + " where c.id = '%d'"
	selectCommentsByTopicID           = selectComments + " where c.topic_id = '%d' order by c.is_pinned desc, c.created_at asc"
	selectCommentsByTopicIDWithOffset = selectComments + " where c.topic_id = '%d' and c.created_at > '%d' order by c.is_pinned desc, c.created_at asc"

	insertComment    = "INSERT INTO comments (topic_id, user_id, message, created_at, user_ip, user_agent, is_pinned, is_deleted) VALUES (:c.topic_id, :c.user_id, :c.message, :c.created_at, :c.user_ip, :c.user_agent, :c.is_pinned, :c.is_deleted)"
	inserTopic       = "INSERT INTO topics (type, board_id, user_id, subject, message, created_at, bumped_at, user_ip, user_agent, is_closed, is_pinned, is_deleted, allow_attach, only_anonymously) VALUES (:t.type, :t.board_id, :t.user_id, :t.subject, :t.message, :t.created_at, :t.bumped_at, :t.user_ip, :t.user_agent, :t.is_closed, :t.is_pinned, :t.is_deleted, :t.allow_attach, :t.only_anonymously)"
	inserUser        = "INSERT INTO users (username, password, created_at, is_banned, is_deleted) VALUES (:u.username, :u.password, :u.created_at, :u.is_banned, :u.is_deleted)"
	inserUserProfile = "INSERT INTO users_profile (user_id, screen_name) VALUES (:u.id, :up.screen_name)"
	inserUserStats   = "INSERT INTO users_stats (user_id) values (:u.id)"

	updateTopicBumpTime = "UPDATE topics as t SET t.bumped_at = '%d' WHERE t.id = '%d'"
)

// NewStorage - init new storage
func NewStorage() *Storage {
	db, err := sqlx.Connect("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Hour)
	// Unsafe becouse i sleep
	return &Storage{db.Unsafe()}
}

// BeginTx - Start transaction
func (s *Storage) BeginTx() {
	// s.db.BeginTransaction()
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
	sql := selectBoards

	err := s.db.Select(&boards, sql)
	if err != nil {
		return nil, err
	}

	return boards, nil
}

// GetBoardBySlug - Get boarf by slug
func (s *Storage) GetBoardBySlug(slug string) (*Board, error) {
	board := Board{}
	sql := fmt.Sprintf(selectBoardBySlug, slug)

	err := s.db.Get(&board, sql)
	if err != nil {
		return nil, err
	}

	return &board, nil
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
	Slug  string
	Sort  string
	Page  int64
	Limit int64
}

// Bind - Bind HTTP request data and validate it
func (tr *TopicsRequest) Bind(r *http.Request) error {

	if slug := r.URL.Query().Get("slug"); slug != "" {
		tr.Slug = utils.EscapeString(slug)
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if pageInt, err := strconv.ParseInt(page, 10, 64); err == nil {
			tr.Page = utils.LimitMinValue(utils.Abs(pageInt), 1)
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if limitInt, err := strconv.ParseInt(limit, 10, 64); err == nil {
			tr.Limit = utils.LimitMaxValue(utils.Abs(limitInt), 64)
		}
	}

	return nil
}

// GetTopicsList - Return topics list with params
func (s *Storage) GetTopicsList(request *TopicsRequest) ([]*Topic, error) {
	topics := []*Topic{}
	sql := selectTopics

	if len(request.Slug) > 0 {
		sql = sql + " " + fmt.Sprintf("where b.slug = '%s'", request.Slug)
	}

	limit := request.Limit
	offset := request.Limit * (request.Page - 1)

	sql = sql + " " + fmt.Sprintf("group by t.id order by t.is_pinned desc, %s desc limit %d offset %d", request.Sort, limit, offset)

	err := s.db.Select(&topics, sql)
	if err != nil {
		return nil, err
	}

	for _, topic := range topics {
		if topic.FilesCount > 0 {
			topic.Attachments = s.GetTopicFiles(topic)
		} else {
			topic.Attachments = []*File{}
		}
	}

	return topics, nil
}

// GetTopicByID - Return topic by ID
func (s *Storage) GetTopicByID(id int64) (*Topic, error) {
	topic := Topic{}
	sql := selectTopics

	if id != 0 {
		sql = sql + " " + fmt.Sprintf("where t.id = '%d'", id)
	}

	sql = sql + " group by t.id"

	err := s.db.Get(&topic, sql)
	if err != nil {
		return nil, err
	}

	topic.Attachments = s.GetTopicFiles(&topic)

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

// UpdateTopicBumpTime - Update topic bump time with comment data
func (s *Storage) UpdateTopicBumpTime(request *Comment) error {
	sql := fmt.Sprintf(updateTopicBumpTime, request.CreatedAt, request.TopicID)

	_, err := s.db.Exec(sql)

	return err
}

// GetTopicFiles - Возвращает файлы топика
// TODO: Ну что за фигня. Надо сделать проще
func (s *Storage) GetTopicFiles(topic *Topic) []*File {
	files := []*File{}
	sql := fmt.Sprintf("select f.* from topics_files as tf left join files as f on tf.file_id = f.id where tf.topic_id = '%d' group by tf.file_id", topic.ID)

	err := s.db.Select(&files, sql)
	if err != nil {
		return nil
	}

	return files
}

// --
// Comments methods
// --

// CommentsRequest - Request for fetch comments
type CommentsRequest struct {
	TopicID int
	Offset  int // Смещение по времени комментария
}

// Bind - Bind HTTP request data and validate it
func (cr *CommentsRequest) Bind(r *http.Request) error {

	if topicID := chi.URLParam(r, "topicID"); topicID != "" {
		if topicIDInt, err := strconv.Atoi(topicID); err == nil {
			cr.TopicID = topicIDInt
		}
	}

	if cr.TopicID == 0 {
		return errors.New("ID needed")
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil {
			cr.Offset = offsetInt
		}
	}

	return nil
}

// GetCommentsList - Return list of comments by topic ID
func (s *Storage) GetCommentsList(request *CommentsRequest) ([]*Comment, error) {
	comments := []*Comment{}
	sql := fmt.Sprintf(selectCommentsByTopicIDWithOffset, request.TopicID, request.Offset)

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
	sql := selectUsers + " " + fmt.Sprintf("where u.username = '%s'", username)

	err := s.db.Get(&user, sql)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID - Return user by ID
func (s *Storage) GetUserByID(id int64) (*User, error) {
	user := User{}
	sql := selectUsers + " " + fmt.Sprintf("where u.id = '%d'", id)

	err := s.db.Get(&user, sql)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser - Create user and return him, or error
// TODO: транзакции
func (s *Storage) CreateUser(request *User) (*User, error) {
	result, err := s.db.NamedExec(inserUser, request)
	if err != nil {
		return nil, err
	}

	UserID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// And we must create profile and stats to user...
	// TODO: Transfer to user controller? Like comment bump

	request.ID = UserID

	if _, err := s.db.NamedExec(inserUserProfile, request); err != nil {
		return nil, err
	}

	if _, err := s.db.NamedExec(inserUserStats, request); err != nil {
		return nil, err
	}

	user, err := s.GetUserByID(UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserStatistic - Get user stats
func (s *Storage) GetUserStatistic(id int64) (*UserStatistic, error) {
	statistic := UserStatistic{}
	sql := selectUsersStatistic + " " + fmt.Sprintf("where us.user_id = '%d'", id)

	err := s.db.Get(&statistic, sql)
	if err != nil {
		return nil, err
	}

	return &statistic, nil
}

// UpdateUserStatistic - Update user stats fields
func (s *Storage) UpdateUserStatistic(id int64, field string) error {
	sql := fmt.Sprintf("UPDATE users_stats as us SET us.%s = us.%s + 1 WHERE us.user_id = '%d'", field, field, id)

	_, err := s.db.Exec(sql)

	return err
}
