package repository

import (
	"github.com/jackc/pgx"
	"techpark_db/internal/pkg/models"
)

type Repository interface {
	Status() (*models.Status, error)
	Clear() error
}

type DBServiceRepository struct {
	Conn *pgx.ConnPool
}

func NewDBServiceRepository(conn *pgx.ConnPool) *DBServiceRepository {
	return &DBServiceRepository{Conn: conn}
}

func (db DBServiceRepository) Status() (*models.Status, error) {
	var forums, posts, threads, users int64

	db.Conn.QueryRow("select count(*) from forums").Scan(&forums)
	db.Conn.QueryRow("select count(*) from posts").Scan(&posts)
	db.Conn.QueryRow("select count(*) from threads").Scan(&threads)
	db.Conn.QueryRow("select count(*) from users").Scan(&users)

	return &models.Status{
		Forums:  forums,
		Posts:   posts,
		Threads: threads,
		Users:   users,
	}, nil
}

func (db DBServiceRepository) Clear() error {
	query := "truncate table votes RESTART IDENTITY cascade;truncate table posts RESTART IDENTITY cascade ;truncate table forums RESTART IDENTITY cascade ;truncate table threads RESTART IDENTITY cascade ;truncate table users RESTART IDENTITY cascade ;"

	_, err := db.Conn.Exec(query)
	return err
}
