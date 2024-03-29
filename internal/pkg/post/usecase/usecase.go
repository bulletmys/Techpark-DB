package usecase

import (
	"techpark_db/internal/pkg/forum"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/post"
	"techpark_db/internal/pkg/thread"
	"techpark_db/internal/pkg/user"
	"time"
)

type PostUC struct {
	PostRepo   post.Repository
	UserRepo   user.Repository
	ThreadRepo thread.Repository
	ForumRepo  forum.Repository
}

func (uc PostUC) Create(posts []*models.Post, slug string, id int32) error {
	threadID, forum := uc.ThreadRepo.FindAndGetID(slug, id)
	if threadID == -1 {
		return models.ThreadNotFound
	}

	if len(posts) == 0 {
		return nil
	}

	created := time.Now().Truncate(time.Microsecond)

	for _, elem := range posts {
		elem.Thread = threadID
		elem.Created = created
		elem.IsEdited = false
		elem.Forum = forum
		dbUser, err := uc.UserRepo.FindUserByNickname(elem.Author)
		if err != nil {
			return err
		}
		if dbUser == nil {
			return models.UserNotFound
		}
	}

	if err := uc.PostRepo.FindPostsByID(posts); err != nil {
		return models.PostNotFound
	}
	return uc.PostRepo.Create(posts)
}

func (uc PostUC) Find(slug string, threadId, limit int32, since int64, desc bool, sortType post.SortType) ([]models.Post, error) {
	var dbThread *models.Thread
	var err error
	if threadId != -1 {
		dbThread, err = uc.ThreadRepo.FindThreadByID(threadId)
	} else {
		dbThread, err = uc.ThreadRepo.FindThreadBySlug(slug)
	}
	if err != nil {
		return nil, err
	}
	if dbThread == nil {
		return nil, models.ThreadNotFound
	}

	var posts []models.Post

	switch sortType {
	case post.FLAT:
		posts, err = uc.PostRepo.FindPostsFlatSort(dbThread.ID, limit, since, desc)
	case post.TREE:
		posts, err = uc.PostRepo.FindPostsTreeSort(dbThread.ID, limit, since, desc)
	case post.PARENT_TREE:
		posts, err = uc.PostRepo.FindPostsParentTreeSort(dbThread.ID, limit, since, desc)
	default:
		posts, err = uc.PostRepo.FindPostsFlatSort(dbThread.ID, limit, since, desc)
	}

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (uc PostUC) FullPostInfo(id int64, userInfo, forumInfo, threadInfo bool) (*models.Details, error) {
	dbPost, err := uc.PostRepo.GetPost(id)
	if err != nil {
		return nil, err
	}
	if dbPost == nil {
		return nil, models.PostNotFound
	}

	details := models.Details{Post: *dbPost}

	if userInfo {
		dbUser, err := uc.UserRepo.FindUserByNickname(dbPost.Author)
		if err != nil {
			return nil, err
		}
		if dbUser == nil {
			return nil, models.UserNotFound
		}
		details.Author = dbUser
	}

	if threadInfo {
		dbThread, err := uc.ThreadRepo.FindThreadByID(dbPost.Thread)
		if err != nil {
			return nil, err
		}
		if dbThread == nil {
			return nil, models.ThreadNotFound
		}
		details.Thread = dbThread
	}

	if forumInfo {
		dbForum, err := uc.ForumRepo.FindForumBySlug(dbPost.Forum)
		if err != nil {
			return nil, err
		}
		if dbForum == nil {
			return nil, models.ForumNotFound
		}
		details.Forum = dbForum
	}
	return &details, nil
}

func (uc PostUC) Update(id int64, msg string) (*models.Post, error) {
	dbPost, err := uc.PostRepo.GetPost(id)
	if err != nil {
		return nil, err
	}
	if dbPost == nil {
		return nil, models.PostNotFound
	}
	if msg != "" && msg != dbPost.Message {
		dbPost.Message = msg
		dbPost.IsEdited = true
		err = uc.PostRepo.UpdatePost(id, msg)
		if err != nil {
			return nil, err
		}
	}
	return dbPost, nil
}
