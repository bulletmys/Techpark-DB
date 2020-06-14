package thread

import (
	"techpark_db/internal/pkg/models"
)

type Repository interface {
	Create(thread models.Thread) (int32, error)
	Update(id int32, msg, title string) error
	FindAndGetID(slug string, id int32) (int32, string)
	Vote(vote models.Vote, thread *models.Thread) error
	FindBySlugOrID(slug string, id int32) (*models.Thread, error)
	FindThreadBySlug(slug string) (*models.Thread, error)
	FindAndGetIDByID(id int32) (int32, string)
	FindAndGetIDBySlug(slug string) (int32, string)
	FindThreadByID(id int32) (*models.Thread, error)
	GetThreadsByForum(forumSlug string, limit int, since string, desc bool) ([]models.Thread, error)
}
