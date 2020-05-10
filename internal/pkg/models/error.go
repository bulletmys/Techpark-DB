package models

import "github.com/pkg/errors"

var SameUserExists = errors.New("user with same data is already exists")

var UserNotFound = errors.New("can't find user with this data")
