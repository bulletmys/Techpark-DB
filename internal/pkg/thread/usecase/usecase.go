package usecase

import (
	"fmt"
	"techpark_db/internal/pkg/forum"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/thread"
	"techpark_db/internal/pkg/user"
)

type ThreadUC struct {
	ThreadRepo thread.Repository
	ForumRepo  forum.Repository
	UserRepo   user.Repository
}

func (uc ThreadUC) Create(thread *models.Thread) error {
	dbUser, err := uc.UserRepo.FindUserByNickname(thread.Author)
	if err != nil {
		return err
	}
	if dbUser == nil {
		return models.UserNotFound
	}

	dbForum, err := uc.ForumRepo.FindForumBySlug(thread.Forum)
	if err != nil {
		return err
	}
	if dbForum == nil {
		return models.ForumNotFound
	}

	thread.Forum = dbForum.Slug

	dbThread, err := uc.ThreadRepo.FindThreadBySlug(thread.Slug)
	if err != nil {
		return err
	}
	if dbThread != nil {
		*thread = *dbThread
		return models.SameThreadExists
	}

	id, err := uc.ThreadRepo.Create(*thread)
	if err != nil {
		return err
	}

	thread.ID = id

	return nil
}

func (uc ThreadUC) GetForumsThreads(forumSlug string, limit int, since string, desc bool) ([]models.Thread, error) {
	dbForum, err := uc.ForumRepo.FindForumBySlug(forumSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to find forum by slug: %v", err)
	}
	if dbForum == nil {
		return nil, models.ForumNotFound
	}

	forumSlug = dbForum.Slug

	return uc.ThreadRepo.GetThreadsByForum(forumSlug, limit, since, desc)
}

func (uc ThreadUC) Vote(vote models.Vote, slug string, id int32) (*models.Thread, error) {
	dbUser, err := uc.UserRepo.FindUserByNickname(vote.Nick)
	if err != nil {
		return nil, err
	}
	if dbUser == nil {
		return nil, models.UserNotFound
	}
	dbThread, err := uc.ThreadRepo.FindBySlugOrID(slug, id)
	if err != nil {
		return nil, err
	}
	if dbThread == nil {
		return nil, models.ThreadNotFound
	}
	//dbThread := *models.Thread{ID: id, Slug: slug}

	return dbThread, uc.ThreadRepo.Vote(vote, dbThread)
}

func (uc ThreadUC) GetThread(slug string, id int32) (*models.Thread, error) {
	dbThread, err := uc.ThreadRepo.FindBySlugOrID(slug, id)
	if err != nil {
		return nil, err
	}
	if dbThread == nil {
		return nil, models.ThreadNotFound
	}
	return dbThread, nil
}

func (uc ThreadUC) Update(id int32, slug, msg, title string) (*models.Thread, error) {
	dbThread, err := uc.ThreadRepo.FindBySlugOrID(slug, id)
	if err != nil {
		return nil, err
	}
	if dbThread == nil {
		return nil, models.ThreadNotFound
	}

	if msg != "" {
		dbThread.Message = msg
	}
	if title != "" {
		dbThread.Title = title
	}

	return dbThread, uc.ThreadRepo.Update(dbThread.ID, msg, title)
}
