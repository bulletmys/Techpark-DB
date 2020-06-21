package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"techpark_db/internal/pkg/models"
)

type Repository interface {
	Status() (*models.Status, error)
	Clear() error
}

type DBServiceRepository struct {
	Conn *pgxpool.Pool
}

func NewDBServiceRepository(conn *pgxpool.Pool) *DBServiceRepository {
	return &DBServiceRepository{Conn: conn}
}

func (db DBServiceRepository) Status() (*models.Status, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var forums, posts, threads, users int64

	conn.QueryRow(context.Background(), "select count(*) from forums").Scan(&forums)
	conn.QueryRow(context.Background(), "select count(*) from posts").Scan(&posts)
	conn.QueryRow(context.Background(), "select count(*) from threads").Scan(&threads)
	conn.QueryRow(context.Background(), "select count(*) from users").Scan(&users)

	return &models.Status{
		Forums:  forums,
		Posts:   posts,
		Threads: threads,
		Users:   users,
	}, nil
}

func (db DBServiceRepository) Clear() error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	query := "truncate table votes RESTART IDENTITY cascade;truncate table posts RESTART IDENTITY cascade ;truncate table forums RESTART IDENTITY cascade ;truncate table threads RESTART IDENTITY cascade ;truncate table users RESTART IDENTITY cascade;truncate table forum_users RESTART IDENTITY cascade;"

	_, err = conn.Exec(context.Background(), query)
	return err
}
