package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"strings"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgxpool.Pool
}

func newDBRepository(conn *pgxpool.Pool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) GetUserNick(nick string) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()
	var str string

	err = conn.QueryRow(context.Background(),
		"select nick from users where nick = $1",
		strings.ToLower(nick),
	).Scan(&str)

	return err
}

func (db DBRepository) FindUserByNickname(nick string) (*models.User, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()
	var userModel models.User

	err = conn.QueryRow(context.Background(),
		"select nick, name, email, about from users where nick = $1",
		strings.ToLower(nick),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &userModel, err
}

func (db DBRepository) FindUser(user models.User) ([]models.User, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(),
		"select nick, name, email, about from users where lower(nick) = $1 or lower(email) = $2",
		strings.ToLower(user.Nickname),
		strings.ToLower(user.Email),
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]models.User, 0, 2)

	defer rows.Close()

	//noinspection GoNilness
	for rows.Next() {
		userModel := models.User{}
		if err := rows.Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About); err != nil {
			return nil, err
		}
		users = append(users, userModel)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users, nil
}

func (db DBRepository) Create(user models.User) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(),
		"insert into users(nick, name, email, about) values($1, $2, $3, $4)",
		user.Nickname,
		user.FullName,
		user.Email,
		user.About,
	)

	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}
	return nil
}

func (db DBRepository) FindUserByEmail(email string) (*models.User, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var userModel models.User

	err = conn.QueryRow(context.Background(),
		"select nick, name, email, about from users where lower(email) = $1",
		strings.ToLower(email),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &userModel, err
}

func (db DBRepository) UpdateUser(user models.User) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(),
		"update users set name = $1, email = $2, about = $3 where lower(nick) = $4",
		user.FullName,
		user.Email,
		user.About,
		strings.ToLower(user.Nickname),
	)

	return err
}

func (db DBRepository) GetForumUsers(slug, since string, limit int, desc bool) ([]models.User, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	args := make([]interface{}, 1)
	args[0] = strings.ToLower(slug)

	query := "select * " +
		"from (select u.nick, u.name, u.email, u.about " +
		"from forums f " +
		"full join threads t on f.slug = t.forum " +
		"full join posts p on f.slug = p.forum " +
		"join users u on (t.nick = u.nick or p.nick = u.nick) " +
		"where lower(f.slug) = $1 "

	if since != "" {
		args = append(args, since)
		if desc {
			query += " and lower(u.nick) < lower($2)"
		} else {
			query += " and lower(u.nick) > lower($2)"
		}
	}
	query += " group by u.nick) as ftpu order by lower(nick)"

	if desc {
		query += " desc "
	}

	if limit > 0 {
		query += "limit " + strconv.Itoa(limit)
	}

	rows, err := conn.Query(context.Background(),
		query,
		args...
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	users := make([]models.User, 0)

	for rows.Next() {
		userModel := models.User{}
		if err := rows.Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About); err != nil {
			return nil, err
		}
		users = append(users, userModel)
	}

	if len(users) == 0 {
		return []models.User{}, nil
	}

	return users, nil
}
