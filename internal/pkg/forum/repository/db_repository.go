package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgxpool.Pool
}

func newDBRepository(conn *pgxpool.Pool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) Create(user models.Forum) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(),
		"insert into forums(slug, nick, title) values($1, $2, $3)",
		user.Slug,
		user.User,
		user.Title,
	)

	if err != nil {
		return fmt.Errorf("failed to insert forum: %v", err)
	}
	return nil
}

func (db DBRepository) FindForum(forum models.Forum) (*models.Forum, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var forumModel models.Forum

	err = conn.QueryRow(context.Background(),
		"select slug, nick, title, threads, posts from forums where nick = $1 and slug = $2",
		strings.ToLower(forum.User),
		strings.ToLower(forum.Slug),
	).Scan(&forumModel.Slug, &forumModel.User, &forumModel.Title, &forumModel.Threads, &forumModel.Posts)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &forumModel, err
}

func (db DBRepository) FindForumBySlug(slug string) (*models.Forum, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var forumModel models.Forum

	err = conn.QueryRow(context.Background(),
		"select slug, nick, title, threads, posts from forums where slug = $1",
		strings.ToLower(slug),
	).Scan(&forumModel.Slug, &forumModel.User, &forumModel.Title, &forumModel.Threads, &forumModel.Posts)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &forumModel, err
}
