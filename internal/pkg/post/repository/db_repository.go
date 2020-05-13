package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"techpark_db/internal/pkg/models"
)

type DBRepository struct {
	Conn *pgxpool.Pool
}

func newDBRepository(conn *pgxpool.Pool) *DBRepository {
	return &DBRepository{Conn: conn}
}

func (db DBRepository) Create(posts []*models.Post) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	for _, elem := range posts {
		var id int64
		err := conn.QueryRow(context.Background(),
			"insert into posts(nick, message, parent, thread, forum, isEdited, created) values($1, $2, $3, $4, $5, $6, $7) returning id",
			elem.Author,
			elem.Message,
			elem.Parent,
			elem.Thread,
			elem.Forum,
			elem.IsEdited,
			elem.Created,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to create posts: %v", err)
		}
		elem.ID = id
		fmt.Printf("%+v\n", elem)
	}

	return nil
}

func (db DBRepository) FindPostsByID(posts []*models.Post) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	for _, elem := range posts {
		if elem.Parent == 0 {
			continue
		}
		var dbID int64
		err := conn.QueryRow(context.Background(),
			"select id from posts where id = $1 and thread = $2",
			elem.Parent,
			elem.Thread,
		).Scan(&dbID)
		if err != nil {
			return fmt.Errorf("failed to find posts: %v", err)
		}
	}

	return nil
}

func configParentTreeQuery(limit int, since int64, desc bool, flag *bool) string {
	query := "select nick, created, forum, id, message, thread, parent from posts where thread = $1 and path[1] = id"

	if since > 0 {
		*flag = true
		if desc {
			query += " and id < $2"
		} else {
			query += " and id > $2"
		}
	}

	query += " order by path"

	if desc {
		query += " desc"
	}

	query += ", id"

	if limit > 0 {
		query += " limit " + strconv.Itoa(int(limit))
	}
	return query
}

func (db DBRepository) FindPostsParentTree(thread, limit int32, since int64, desc bool) ([]models.Post, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	query := "select nick, created, forum, id, message, thread, parent from get_parent_tree($1, $2, $3)"

	args := make([]interface{}, 3)
	args[0] = limit
	args[1] = thread

	if desc {
		args[0] = thread
		query = "SELECT p2.nick, p2.created, p2.forum, p2.id, p2.message, p2.thread, p2.parent FROM (select * from posts WHERE parent = 0 and thread = $1 order by id desc "
		if since > 0 {
			query = fmt.Sprintf("with sincePost AS (select * from posts where id = %s)SELECT p2.nick, p2.created, p2.forum, p2.id, p2.message, p2.thread, p2.parent FROM (select * from posts WHERE parent = 0 and thread = $1 and id < (select sincePost.path[1] from sincePost) order by id desc ", strconv.FormatInt(since, 10))
		}
		if limit > 0 {
			args[1] = limit
			query += "limit $2"
		}
		query += ") p1 join posts p2 on (p1.id = p2.path[1] or p1.id = p2.id) order by p1.id desc, p2.path offset $3;"
	}

	var offset = 0
	if since > 0 && !desc {
		err = conn.QueryRow(context.Background(), "SELECT get_all_foo($1, $2)", since, thread).Scan(&offset)
		if err != nil {
			return nil, fmt.Errorf("failed to set config posts: %v", err)
		}
	}
	args[2] = offset

	rows, err := conn.Query(context.Background(),
		query,
		args...
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find posts: %v", err)
	}

	posts := make([]models.Post, 0)
	defer rows.Close()

	for rows.Next() {
		postModel := models.Post{}
		if err := rows.Scan(
			&postModel.Author,
			&postModel.Created,
			&postModel.Forum,
			&postModel.ID,
			&postModel.Message,
			&postModel.Thread,
			&postModel.Parent,
		); err != nil {
			return nil, fmt.Errorf("error while scaning query rows: %v", err)
		}

		posts = append(posts, postModel)
	}

	return posts, nil
}

func (db DBRepository) FindPostsFlat(thread, limit int32, since int64, desc bool, isTree bool) ([]models.Post, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	flag := false
	query := findPostsQueryConfigurator(int(limit), since, desc, &flag, isTree)

	args := make([]interface{}, 1)
	args[0] = thread

	if flag {
		args = append(args, since)
	}

	var offset int
	if isTree && since > 0 {
		qry := "SELECT get_all_foo"
		if desc {
			qry += "2"
		}
		qry += "($1, $2)"
		err = conn.QueryRow(context.Background(), qry, since, thread).Scan(&offset)
		if err != nil {
			return nil, fmt.Errorf("failed to set config posts: %v", err)
		}
	}

	query += " offset " + strconv.Itoa(offset)

	fmt.Println(query)

	rows, err := conn.Query(context.Background(),
		query,
		args...
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find posts: %v", err)
	}

	posts := make([]models.Post, 0)
	defer rows.Close()

	for rows.Next() {
		postModel := models.Post{}
		if err := rows.Scan(
			&postModel.Author,
			&postModel.Created,
			&postModel.Forum,
			&postModel.ID,
			&postModel.Message,
			&postModel.Thread,
			&postModel.Parent,
		); err != nil {
			return nil, fmt.Errorf("error while scaning query rows: %v", err)
		}

		posts = append(posts, postModel)
	}
	return posts, nil
}

func findPostsQueryConfigurator(limit int, since int64, desc bool, flag *bool, isTree bool) string {
	query := "select nick, created, forum, id, message, thread, parent from posts where thread = $1"

	if since > 0 {
		if !isTree {
			*flag = true
			if desc {
				query += " and id < $2"
			} else {
				query += " and id > $2"
			}
		}
	}

	if isTree {
		query += " order by path"
	} else {
		query += " order by created"
	}

	if desc {
		query += " desc"
	}

	query += ", id"

	if desc {
		query += " desc"
	}

	if limit > 0 {
		query += " limit " + strconv.Itoa(limit)
	}

	return query
}

func configFindPostsQuery(limit int, since int64, desc bool, flag *bool) string {
	query := "select nick, created, forum, id, message, thread, parent from posts where thread = $1"

	if since > 0 {
		*flag = true
		if desc {
			query += " and id < $2"
		} else {
			query += " and id > $2"
		}
	}

	query += " order by id"

	if desc {
		query += " desc"
	}

	if limit > 0 {
		query += " limit " + strconv.Itoa(int(limit))
	}
	return query
}

func (db DBRepository) FindPosts(thread, limit int32, since int64, desc bool) ([]models.Post, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	flag := false

	query := configFindPostsQuery(int(limit), since, desc, &flag)

	args := make([]interface{}, 1)
	args[0] = thread

	if flag {
		args = append(args, since)
	}

	rows, err := conn.Query(context.Background(),
		query,
		args...
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find posts: %v", err)
	}

	posts := make([]models.Post, 0)
	defer rows.Close()

	for rows.Next() {
		postModel := models.Post{}
		if err := rows.Scan(
			&postModel.Author,
			&postModel.Created,
			&postModel.Forum,
			&postModel.ID,
			&postModel.Message,
			&postModel.Thread,
			&postModel.Parent,
		); err != nil {
			return nil, fmt.Errorf("error while scaning query rows: %v", err)
		}

		posts = append(posts, postModel)
	}

	return posts, nil
}

func (db DBRepository) GetPost(id int64) (*models.Post, error) {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	var post models.Post

	query := "select nick, created, forum, id, isEdited, message, thread, parent from posts where id = $1"

	err = conn.QueryRow(context.Background(),
		query,
		id,
	).Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.IsEdited, &post.Message, &post.Thread, &post.Parent)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %v", err)
	}
	return &post, nil
}

func (db DBRepository) UpdatePost(id int64, msg string) error {
	conn, err := db.Conn.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire conn: %v", err)
	}
	defer conn.Release()

	query := "update posts set message = $2, isEdited = true where id = $1"

	_, err = conn.Exec(context.Background(),
		query,
		id,
		msg,
	)
	if err != nil {
		return  fmt.Errorf("failed to update post: %v", err)
	}

	return nil
}
