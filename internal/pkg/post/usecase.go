package post

import (
	"techpark_db/internal/pkg/models"
)

type UseCase interface {
	Create(posts []*models.Post, slug string, id int32) error
	Find(slug string, threadId, limit int32, since int64, desc bool, sortType SortType) ([]models.Post, error)
	FullPostInfo(id int64, userInfo, forumInfo, threadInfo bool) (*models.Details, error)
	Update(id int64, msg string) (*models.Post, error)
}