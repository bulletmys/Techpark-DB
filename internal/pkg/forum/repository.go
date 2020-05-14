package forum

import "techpark_db/internal/pkg/models"

type Repository interface {
	Create(forum models.Forum) error
	FindForum(forum models.Forum) (*models.Forum, error)
	FindForumBySlug(slug string) (*models.Forum, error)
}
