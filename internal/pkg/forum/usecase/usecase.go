package usecase

import (
	"techpark_db/internal/pkg/forum"
	"techpark_db/internal/pkg/models"
	"techpark_db/internal/pkg/user"
)

type ForumUC struct {
	ForumRepo forum.Repository
	UserRepo  user.Repository
}

func (uc ForumUC) Create(forum models.Forum) (*models.Forum, error) {
	dbUser, err := uc.UserRepo.FindUserByNickname(forum.User)
	if err != nil {
		return nil, err
	}
	if dbUser == nil {
		return nil, models.UserNotFound
	}

	dbForum, err := uc.ForumRepo.FindForumBySlug(forum.Slug)
	if err != nil {
		return nil, err
	}
	if dbForum != nil {
		return dbForum, models.SameForumExists
	}

	return &forum, uc.ForumRepo.Create(forum)
}

func (uc ForumUC) Find(slug string) (*models.Forum, error) {
	dbForum, err := uc.ForumRepo.FindForumBySlug(slug)
	if err != nil {
		return nil, err
	}
	if dbForum == nil {
		return nil, models.ForumNotFound
	}
	return dbForum, nil
}
