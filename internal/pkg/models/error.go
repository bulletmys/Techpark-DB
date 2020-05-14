package models

import "github.com/pkg/errors"

var SameUserExists = errors.New("user with same data is already exists")

var UserNotFound = errors.New("can't find user with this data")

var ForumNotFound = errors.New("can't find forum with this data")

var SameForumExists = errors.New("forum with same data is already exists")

var ThreadNotFound = errors.New("can't find thread with this data")

var SameThreadExists = errors.New("thread with same data is already exists")

var PostNotFound = errors.New("can't find post with this data")

var SamePostExists = errors.New("post with same data is already exists")
