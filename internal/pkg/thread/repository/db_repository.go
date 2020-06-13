package repository

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx"
	"strconv"
	"strings"
	"techpark_db/internal/pkg/models"
	"time"
)

type DBRepository struct {
	Conn *pgx.ConnPool
}

func newDBRepository(conn *pgx.ConnPool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) Create(thread models.Thread) (int32, error) {
	var id int32

	fields := make([]interface{}, 0)
	fields = append(fields, thread.Author, thread.Forum, thread.Title, thread.Message)

	queryParams := "nick, forum, title, message"
	queryValues := "$1, $2, $3, $4"
	counter := 4

	if thread.Slug != "" {
		queryParams += ", slug"
		counter++
		queryValues += ",$" + strconv.Itoa(counter)
		fields = append(fields, thread.Slug)
	}
	if !thread.Created.IsZero() {
		queryParams += ", created"
		counter++
		queryValues += ",$" + strconv.Itoa(counter)
		fields = append(fields, thread.Created)
	}

	query := fmt.Sprintf("insert into threads(%s) values(%s) RETURNING id", queryParams, queryValues)

	err := db.Conn.QueryRow(
		query,
		fields...
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert thread: %v", err)
	}
	return id, nil
}

func (db DBRepository) FindThreadBySlug(slug string) (*models.Thread, error) {
	var thread models.Thread

	err := db.Conn.QueryRow(
		"select nick, created, forum, id, message, slug, title, votes from threads where slug = $1",
		strings.ToLower(slug),
	).Scan(
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.ID,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Votes,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find thread by slug: %v", err)
	}

	return &thread, nil
}

func (db DBRepository) FindThreadByID(id int32) (*models.Thread, error) {
	var thread models.Thread

	err := db.Conn.QueryRow(
		"select nick, created, forum, id, message, slug, title, votes from threads where id = $1",
		id,
	).Scan(
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.ID,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Votes,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find thread by id: %v", err)
	}

	return &thread, nil
}

func (db DBRepository) FindAndGetIDBySlug(slug string) (int32, string) {
	var id int32
	var forum string
	err := db.Conn.QueryRow(
		"select id, forum from threads where slug = $1",
		strings.ToLower(slug),
	).Scan(&id, &forum)
	if err != nil {
		return -1, ""
	}
	return id, forum
}

func (db DBRepository) FindAndGetIDByID(id int32) (int32, string) {
	var forum string
	err := db.Conn.QueryRow(
		"select id, forum from threads where id = $1",
		id,
	).Scan(&id, &forum)
	if err != nil {
		return -1, ""
	}
	return id, forum
}

func (db DBRepository) FindAndGetID(slug string, id int32) (int32, string) {

	var forum string
	err := db.Conn.QueryRow(
		"select id, forum from threads where slug = $1 or id = $2",
		strings.ToLower(slug),
		id,
	).Scan(&id, &forum)
	if err != nil {
		return -1, ""
	}
	return id, forum
}

func (db DBRepository) FindBySlugOrID(slug string, id int32) (*models.Thread, error) {
	var thread models.Thread

	var sluggg sql.NullString

	err := db.Conn.QueryRow(
		"select nick, created, forum, id, message, title, votes, slug from threads where id = $1 or slug = $2",
		id,
		strings.ToLower(slug),
	).Scan(
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.ID,
		&thread.Message,
		&thread.Title,
		&thread.Votes,
		&sluggg,
	)
	if sluggg.Valid {
		thread.Slug = sluggg.String
	}
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (db DBRepository) GetThreadsByForum(forumSlug string, limit int, since time.Time, desc bool) ([]models.Thread, error) {
	flag := false
	query := getThreadsQueryConfigurator(limit, since, desc, &flag)

	args := make([]interface{}, 1)
	args[0] = forumSlug

	if flag {
		args = append(args, since)
	}
	rows, err := db.Conn.Query(
		query,
		args...
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %v", err)
	}

	threads := make([]models.Thread, 0)

	defer rows.Close()

	for rows.Next() {
		threadModel := models.Thread{}
		if err := rows.Scan(
			&threadModel.Author,
			&threadModel.Created,
			&threadModel.Forum,
			&threadModel.ID,
			&threadModel.Message,
			&threadModel.Slug,
			&threadModel.Title,
			&threadModel.Votes,
		); err != nil {
			return nil, fmt.Errorf("error while scaning query rows: %v", err)
		}

		threads = append(threads, threadModel)
	}

	return threads, nil
}

func getThreadsQueryConfigurator(limit int, since time.Time, desc bool, flag *bool) string {
	query := "select nick, created, forum, id, message, slug, title, votes from threads where forum = $1"

	if !since.IsZero() {
		*flag = true
		if desc {
			query += " and created <= $2"
		} else {
			query += " and created >= $2"
		}
	}

	query += "order by created"

	if desc {
		query += " desc"
	}

	if limit > 0 {
		query += " limit " + strconv.Itoa(limit)
	}

	return query
}

func (db DBRepository) Vote(vote models.Vote, thread *models.Thread) error {
	userVoice := false
	if vote.Voice > 0 {
		userVoice = true
	}
	var userVote bool
	queryCheck := "select vote from votes where nick = $1 and thread = $2"

	err := db.Conn.QueryRow(
		queryCheck,
		strings.ToLower(vote.Nick),
		thread.ID,
	).Scan(&userVote)

	if err == nil && userVote == userVoice {
		return nil
	}

	var diff int32 = 1
	if userVoice != userVote && err != pgx.ErrNoRows {
		diff = 2
		_, err = db.Conn.Exec(
			"update votes set vote = $1 where nick = $2 and thread = $3",
			userVoice,
			strings.ToLower(vote.Nick),
			thread.ID,
		)
		if err != nil {
			return err
		}
	} else {
		queryNewVote := "insert into votes(nick, vote, thread) values ($1, $2, $3)"
		_, err = db.Conn.Exec(
			queryNewVote,
			vote.Nick,
			userVoice,
			thread.ID,
		)
		if err != nil {
			return err
		}
	}

	query := ""
	if !userVoice {
		query = fmt.Sprintf("update threads set votes = votes - %v ", diff)
		thread.Votes -= diff
	} else {
		query = fmt.Sprintf("update threads set votes = votes + %v ", diff)
		thread.Votes += diff
	}

	query += "where id = $1"

	_, err = db.Conn.Exec(
		query,
		thread.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update votes in thread: %v", err)
	}
	return nil
}

func (db DBRepository) Update(id int32, msg, title string) error {
	args := make([]interface{}, 1)
	args[0] = id

	query := "update threads set"
	flag := false

	if msg != "" {
		query += " message = $2"
		flag = true
		args = append(args, msg)
	}

	if title != "" {
		if flag {
			query += ", title = $3"
		} else {
			query += " title = $2"
		}
		args = append(args, title)
	}

	if len(args) == 1 {
		return nil
	}

	query += " where id = $1"

	_, err := db.Conn.Exec(
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("failed to update thread: %v", err)
	}

	return nil
}
