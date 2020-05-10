package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strings"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgx.Conn
}

func newDBRepository(conn *pgx.Conn) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) FindUserByNickname(nick string) (*models.User, error) {
	var userModel models.User

	err := db.Conn.QueryRow(context.Background(),
		"select nick, name, email, about from users where lower(nick) = $1",
		strings.ToLower(nick),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &userModel, err
}

func (db DBRepository) FindUser(user models.User) ([]models.User, error) {

	rows, err := db.Conn.Query(context.Background(),
		"select nick, name, email, about from users where lower(nick) = $1 or lower(email) = $2",
		strings.ToLower(user.Nickname),
		strings.ToLower(user.Email),
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	defer rows.Close()

	users := make([]models.User, 0, 2)

	//noinspection GoNilness
	for rows.Next() {
		userModel := models.User{}
		if err := rows.Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About); err != nil {
			return nil, err
		}
		users = append(users, userModel)
	}

	return users, err
}

func (db DBRepository) Create(user models.User) error {
	_, err := db.Conn.Exec(context.Background(),
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

	err := db.Conn.QueryRow(context.Background(),
		"select nick, name, email, about from users where lower(email) = $1",
		strings.ToLower(email),
	).Scan(&userModel.Nickname, &userModel.FullName, &userModel.Email, &userModel.About)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &userModel, err
}

func (db DBRepository) UpdateUser(user models.User) error {
	_, err := db.Conn.Exec(context.Background(),
		"update users set name = $1, email = $2, about = $3 where lower(nick) = $4",
		user.FullName,
		user.Email,
		user.About,
		strings.ToLower(user.Nickname),
	)

	return err
}
