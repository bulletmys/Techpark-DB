package forum

import "techpark_db/internal/pkg/models"

type UseCase interface {
	Create(forum models.Forum) (*models.Forum, error)
	Find(slug string) (*models.Forum, error)
}
