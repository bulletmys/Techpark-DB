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
	var threadID int32
	var f string

	if id != -1 {
		threadID, f = uc.ThreadRepo.FindAndGetIDByID(id)
	} else {
		threadID, f = uc.ThreadRepo.FindAndGetIDBySlug(slug)
	}

	if threadID == -1 {
		return models.ThreadNotFound
	}

	if len(posts) == 0 {
		return nil
	}

	created := time.Now()

	for _, elem := range posts {
		elem.Thread = threadID
		elem.Created = created
		elem.IsEdited = false
		elem.Forum = f
		if dbUser, _ := uc.UserRepo.FindUserByNickname(elem.Author); dbUser == nil {
			return models.UserNotFound
		}
		if err := uc.PostRepo.CheckParentPostByID(elem); err != nil {
			return models.PostNotFound
		}
	}

	return uc.PostRepo.Create(posts)
}

func (uc PostUC) Find(slug string, threadId, limit int32, since int64, desc bool, sortType post.SortType) ([]models.Post, error) {
	dbThread, err := uc.ThreadRepo.FindBySlugOrID(slug, threadId)
	if err != nil {
		return nil, err
	}
	if dbThread == nil {
		return nil, models.ThreadNotFound
	}

	var posts []models.Post

	switch sortType {
	case post.FLAT:
		posts, err = uc.PostRepo.FindPostsFlat(dbThread.ID, limit, since, desc, false)
	case post.TREE:
		posts, err = uc.PostRepo.FindPostsFlat(dbThread.ID, limit, since, desc, true)
	case post.PARENT_TREE:
		posts, err = uc.PostRepo.FindPostsParentTree(dbThread.ID, limit, since, desc)
	default:
		posts, err = uc.PostRepo.FindPosts(dbThread.ID, limit, since, desc)
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
