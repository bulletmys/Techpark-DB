package thread

import (
	"techpark_db/internal/pkg/models"
)

type UseCase interface {
	Create(thread *models.Thread) error
	Vote(vote models.Vote, slug string, id int32) (*models.Thread, error)
	GetForumsThreads(forumSlug string, limit int, since string, desc bool) ([]models.Thread, error)
	GetThread(slug string, id int32) (*models.Thread, error)
	Update(id int32, slug, msg, title string) (*models.Thread, error)
}
