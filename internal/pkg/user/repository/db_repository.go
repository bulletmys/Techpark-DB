package repository

import (
	"fmt"
	pgx2 "github.com/jackc/pgx"
	"log"
	"strconv"
	"strings"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgx2.ConnPool
}

func newDBRepository(conn *pgx2.ConnPool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) FindUserByNickname(nick string) (*models.User, error) {

	var userModel models.User

	err := db.Conn.QueryRow(
		"select nick, name, email, about from users where lower(nick) = $1",
		strings.ToLower(nick),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx2.ErrNoRows {
		return nil, nil
	}

	return &userModel, err
}

func (db DBRepository) FindUser(user models.User) ([]models.User, error) {

	rows, err := db.Conn.Query(
		"select nick, name, email, about from users where nick = $1 or email = $2",
		strings.ToLower(user.Nickname),
		strings.ToLower(user.Email),
	)

	if err == pgx2.ErrNoRows {
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


	_, err := db.Conn.Exec(
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


	var userModel models.User

	err := db.Conn.QueryRow(
		"select nick, name, email, about from users where lower(email) = $1",
		strings.ToLower(email),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx2.ErrNoRows {
		return nil, nil
	}

	log.Println("ErrorDB:", err)

	return &userModel, err
}

func (db DBRepository) UpdateUser(user models.User) error {

	_, err := db.Conn.Exec(
		"update users set name = $1, email = $2, about = $3 where lower(nick) = $4",
		user.FullName,
		user.Email,
		user.About,
		strings.ToLower(user.Nickname),
	)

	return err
}

func (db DBRepository) GetForumUsers(slug, since string, limit int, desc bool) ([]models.User, error) {


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

	rows, err := db.Conn.Query(
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
