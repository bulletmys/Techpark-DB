package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"strings"
	"techpark_db/internal/pkg/models"
	"time"
)

type DBRepository struct {
	Conn *pgxpool.Pool
}

func newDBRepository(conn *pgxpool.Pool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) Create(thread models.Thread) (int32, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

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

	err = conn.QueryRow(context.Background(),
		query,
		fields...
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert thread: %v", err)
	}
	return id, nil
}

func (db DBRepository) FindThreadBySlug(slug string) (*models.Thread, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var thread models.Thread

	err = conn.QueryRow(context.Background(),
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
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var thread models.Thread

	err = conn.QueryRow(context.Background(),
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

func (db DBRepository) FindAndGetID(slug string, id int32) (int32, string) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return -1, ""
	}
	defer conn.Release()

	var forum string
	err = conn.QueryRow(context.Background(),
		"select id, forum from threads where lower(slug) = $1 or id = $2",
		strings.ToLower(slug),
		id,
	).Scan(&id, &forum)
	if err != nil {
		return -1, ""
	}
	return id, forum
}

func (db DBRepository) FindBySlugOrID(slug string, id int32) (*models.Thread, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var thread models.Thread

	err = conn.QueryRow(context.Background(),
		"select nick, created, forum, id, message, slug, title, votes from threads where id = $1 or lower(slug) = $2",
		id,
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
		return nil, err
	}
	return &thread, nil
}

func (db DBRepository) GetThreadsByForum(forumSlug string, limit int, since time.Time, desc bool) ([]models.Thread, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	flag := false
	query := getThreadsQueryConfigurator(limit, since, desc, &flag)

	args := make([]interface{}, 1)
	args[0] = forumSlug

	if flag {
		args = append(args, since)
	}
	rows, err := conn.Query(context.Background(),
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
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	tx, txErr := conn.Begin(context.Background())
	if txErr != nil {
		return txErr
	}
	defer tx.Rollback(context.Background())

	rows, err := tx.Exec(context.Background(), `UPDATE votes SET vote = $1 WHERE thread = $2 AND nick = $3;`, vote.Voice, thread.ID, vote.Nick)
	if rows.RowsAffected() == 0 {
		_, err := tx.Exec(context.Background(), `INSERT INTO votes (nick, thread, vote) VALUES ($1, $2, $3);`, strings.ToLower(vote.Nick), thread.ID, vote.Voice)
		if err != nil {
			return models.UserNotFound
		}
	}
	err = tx.QueryRow(context.Background(), `SELECT votes FROM threads WHERE id = $1`, thread.ID).Scan(&thread.Votes)
	if err != nil {
		return err
	}
	tx.Commit(context.Background())

	return nil
}

func (db DBRepository) Update(id int32, msg, title string) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

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

	_, err = conn.Exec(context.Background(),
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("failed to update thread: %v", err)
	}

	return nil
}
