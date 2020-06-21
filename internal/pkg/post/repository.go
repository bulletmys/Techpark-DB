package post

import (
	"techpark_db/internal/pkg/models"
)

type Repository interface {
	Create(posts []*models.Post) error
	FindPostsByID(posts []*models.Post) error
	FindPostsFlat(thread, limit int32, since int64, desc bool, isTree bool) ([]models.Post, error)
	FindPostsParentTree(thread, limit int32, since int64, desc bool) ([]models.Post, error)
	FindPosts(thread, limit int32, since int64, desc bool) ([]models.Post, error)
	GetPost(id int64) (*models.Post, error)
	UpdatePost(id int64, msg string) error
	FindPostsAlternative(threadID, limit int32, since int64, desc bool, sortType SortType) ([]models.Post, error)
	FindPostsAlternative2(threadID, limit int32, since int64, desc bool, sortType string) ([]models.Post, error)
}
