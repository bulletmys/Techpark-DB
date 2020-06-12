package repository

import (
	"fmt"
	"github.com/jackc/pgx"
	"strings"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgx.ConnPool
}

func newDBRepository(conn *pgx.ConnPool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) Create(user models.Forum) error {

	_, err := db.Conn.Exec(
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

	var forumModel models.Forum

	err := db.Conn.QueryRow(
		"select slug, nick, title, threads, posts from forums where lower(nick) = $1 and lower(slug) = $2",
		strings.ToLower(forum.User),
		strings.ToLower(forum.Slug),
	).Scan(&forumModel.Slug, &forumModel.User, &forumModel.Title, &forumModel.Threads, &forumModel.Posts)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &forumModel, err
}

func (db DBRepository) FindForumBySlug(slug string) (*models.Forum, error) {

	var forumModel models.Forum

	err := db.Conn.QueryRow(
		"select slug, nick, title, threads, posts from forums where lower(slug) = $1",
		strings.ToLower(slug),
	).Scan(&forumModel.Slug, &forumModel.User, &forumModel.Title, &forumModel.Threads, &forumModel.Posts)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &forumModel, err
}
